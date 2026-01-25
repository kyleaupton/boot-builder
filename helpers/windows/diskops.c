#include "diskops.h"
#include "pipe.h"
#include <windows.h>
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <ctype.h>

/*
 * Disk operations implementation for FlashIt privileged helper.
 * format_disk: Formats disk with specified filesystem using diskpart
 * write_iso: Writes ISO image to disk (cleans disk first, then raw writes)
 */

/* Thread-local error message buffer */
#ifdef __GNUC__
static __thread char g_error_message[MAX_ERROR_MSG];
#else
static __declspec(thread) char g_error_message[MAX_ERROR_MSG];
#endif

/* Global cancellation flag (volatile for cross-thread visibility) */
static volatile int g_cancel_flag = 0;

const char* get_last_error_message(void) {
    return g_error_message;
}

void set_last_error_message(const char *fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vsnprintf(g_error_message, sizeof(g_error_message), fmt, args);
    va_end(args);
}

void cancel_current_operation(void) {
    g_cancel_flag = 1;
}

void reset_cancel_flag(void) {
    g_cancel_flag = 0;
}

int is_cancelled(void) {
    return g_cancel_flag;
}

/* Forward declarations */
static int extract_disk_number(const char *device);
static int run_diskpart(const char *script_path);

int send_progress_update(OperationContext *ctx, uint64_t written, uint64_t total) {
    if (ctx == NULL || ctx->pipe == NULL || ctx->pipe == INVALID_HANDLE_VALUE) {
        return -1;
    }

    Response resp;
    make_progress_response(&resp, ctx->request_id, written, total);
    return send_response(ctx->pipe, &resp);
}

/*
 * Build the normalized device path (\\.\PhysicalDriveN) from disk number.
 * Returns 0 on success, -1 on failure.
 */
static int build_device_path(int disk_num, char *out_path, size_t out_size) {
    int written = snprintf(out_path, out_size, "\\\\.\\PhysicalDrive%d", disk_num);
    if (written < 0 || (size_t)written >= out_size) {
        return -1;
    }
    return 0;
}

/* Buffer size for ISO writing (1MB for good performance) */
#define WRITE_BUFFER_SIZE (1024 * 1024)

/* Progress update interval (~10MB) */
#define PROGRESS_INTERVAL (10 * 1024 * 1024)

/* Disk sector size for alignment */
#define SECTOR_SIZE 512

/*
 * Clean the disk by removing all partitions using diskpart.
 * This prepares the disk for raw ISO writing by removing all volumes
 * so nothing can interfere with exclusive raw disk access.
 * Returns 0 on success, -1 on failure.
 */
static int clean_disk(int disk_num) {
    char temp_path[MAX_PATH];
    char script_path[MAX_PATH];
    FILE *script_file;
    int result;

    /* Safety check: refuse to clean disk 0 (system disk) */
    if (disk_num == 0) {
        set_last_error_message("refusing to clean disk 0 (system disk)");
        return -1;
    }

    /* Get temp directory */
    if (GetTempPathA(MAX_PATH, temp_path) == 0) {
        set_last_error_message("failed to get temp path: error %lu", GetLastError());
        return -1;
    }

    /* Create temp script file */
    if (GetTempFileNameA(temp_path, "fic", 0, script_path) == 0) {
        set_last_error_message("failed to create temp file: error %lu", GetLastError());
        return -1;
    }

    /* Write diskpart script */
    script_file = fopen(script_path, "w");
    if (script_file == NULL) {
        DeleteFileA(script_path);
        set_last_error_message("failed to open temp script file for writing");
        return -1;
    }

    fprintf(script_file, "select disk %d\n", disk_num);
    fprintf(script_file, "clean\n");
    fprintf(script_file, "exit\n");
    fclose(script_file);

    /* Run diskpart */
    result = run_diskpart(script_path);
    DeleteFileA(script_path);

    if (result != 0) {
        if (result > 0) {
            set_last_error_message("diskpart clean failed with exit code %d", result);
        }
        /* If result == -1, error message already set by run_diskpart */
        return -1;
    }

    return 0;
}

int write_iso(const char *device, const char *iso_path, OperationContext *ctx) {
    HANDLE hISO = INVALID_HANDLE_VALUE;
    HANDLE hDrive = INVALID_HANDLE_VALUE;
    void *buffer = NULL;
    char device_path[64];
    LARGE_INTEGER file_size;
    uint64_t total_size;
    uint64_t bytes_written = 0;
    uint64_t last_progress = 0;
    DWORD bytes_read, bytes_to_write, written;
    int disk_num;
    int success = 0;

    /* Validate device path */
    if (device == NULL || device[0] == '\0') {
        set_last_error_message("device path is required");
        return -1;
    }

    /* Validate ISO path */
    if (iso_path == NULL || iso_path[0] == '\0') {
        set_last_error_message("ISO path is required");
        return -1;
    }

    /* Reset cancellation flag before starting */
    reset_cancel_flag();

    /* Extract and validate disk number */
    disk_num = extract_disk_number(device);
    if (disk_num < 0) {
        set_last_error_message("invalid device path: %s (expected PhysicalDriveN format)", device);
        return -1;
    }

    /* Safety check: refuse to write to disk 0 (system disk) */
    if (disk_num == 0) {
        set_last_error_message("refusing to write to disk 0 (system disk)");
        return -1;
    }

    /* Clean the disk to remove all partitions.
     * This ensures exclusive access for raw writing by removing all volumes
     * that could interfere (e.g., Explorer trying to access them). */
    if (clean_disk(disk_num) != 0) {
        return -1;  /* Error message already set by clean_disk */
    }

    /* Build normalized device path */
    if (build_device_path(disk_num, device_path, sizeof(device_path)) != 0) {
        set_last_error_message("failed to build device path");
        return -1;
    }

    /* Open source ISO file for reading */
    hISO = CreateFileA(
        iso_path,
        GENERIC_READ,
        FILE_SHARE_READ,
        NULL,
        OPEN_EXISTING,
        FILE_ATTRIBUTE_NORMAL | FILE_FLAG_SEQUENTIAL_SCAN,
        NULL
    );

    if (hISO == INVALID_HANDLE_VALUE) {
        set_last_error_message("failed to open ISO file '%s': error %lu", iso_path, GetLastError());
        return -1;
    }

    /* Get ISO file size for progress reporting */
    if (!GetFileSizeEx(hISO, &file_size)) {
        set_last_error_message("failed to get ISO file size: error %lu", GetLastError());
        goto cleanup;
    }
    total_size = (uint64_t)file_size.QuadPart;

    if (total_size == 0) {
        set_last_error_message("ISO file is empty");
        goto cleanup;
    }

    /* Open physical drive for raw writing.
     * GENERIC_READ is required in addition to GENERIC_WRITE for raw disk I/O.
     * FILE_FLAG_NO_BUFFERING ensures direct writes, FILE_FLAG_WRITE_THROUGH
     * ensures data is flushed to disk. */
    hDrive = CreateFileA(
        device_path,
        GENERIC_READ | GENERIC_WRITE,
        FILE_SHARE_READ | FILE_SHARE_WRITE,
        NULL,
        OPEN_EXISTING,
        FILE_FLAG_NO_BUFFERING | FILE_FLAG_WRITE_THROUGH,
        NULL
    );

    if (hDrive == INVALID_HANDLE_VALUE) {
        set_last_error_message("failed to open device '%s' for writing: error %lu", device_path, GetLastError());
        goto cleanup;
    }

    /* Allocate sector-aligned buffer using VirtualAlloc
     * FILE_FLAG_NO_BUFFERING requires sector-aligned buffer and sizes */
    buffer = VirtualAlloc(NULL, WRITE_BUFFER_SIZE, MEM_COMMIT, PAGE_READWRITE);
    if (buffer == NULL) {
        set_last_error_message("failed to allocate write buffer: error %lu", GetLastError());
        goto cleanup;
    }

    /* Send initial progress update */
    if (ctx != NULL) {
        send_progress_update(ctx, 0, total_size);
    }

    /* Main write loop */
    while (1) {
        /* Check for cancellation before each chunk */
        if (is_cancelled()) {
            set_last_error_message("cancelled: operation was cancelled");
            goto cleanup;
        }

        /* Read a chunk from ISO */
        if (!ReadFile(hISO, buffer, WRITE_BUFFER_SIZE, &bytes_read, NULL)) {
            set_last_error_message("failed to read from ISO file: error %lu", GetLastError());
            goto cleanup;
        }

        /* End of file */
        if (bytes_read == 0) {
            break;
        }

        /* Calculate bytes to write (must be sector-aligned for FILE_FLAG_NO_BUFFERING) */
        bytes_to_write = bytes_read;
        if (bytes_to_write % SECTOR_SIZE != 0) {
            /* Pad to next sector boundary */
            DWORD padding = SECTOR_SIZE - (bytes_to_write % SECTOR_SIZE);
            memset((char*)buffer + bytes_read, 0, padding);
            bytes_to_write += padding;
        }

        /* Write chunk to device */
        if (!WriteFile(hDrive, buffer, bytes_to_write, &written, NULL)) {
            set_last_error_message("failed to write to device: error %lu", GetLastError());
            goto cleanup;
        }

        if (written != bytes_to_write) {
            set_last_error_message("incomplete write: expected %lu bytes, wrote %lu", bytes_to_write, written);
            goto cleanup;
        }

        /* Track actual ISO bytes written (not padded amount) */
        bytes_written += bytes_read;

        /* Send progress update periodically */
        if (ctx != NULL && (bytes_written - last_progress >= PROGRESS_INTERVAL)) {
            send_progress_update(ctx, bytes_written, total_size);
            last_progress = bytes_written;
        }
    }

    /* Flush device buffers to ensure all data is written */
    FlushFileBuffers(hDrive);

    /* Send final progress update */
    if (ctx != NULL) {
        send_progress_update(ctx, bytes_written, total_size);
    }

    success = 1;

cleanup:
    if (buffer != NULL) {
        VirtualFree(buffer, 0, MEM_RELEASE);
    }
    if (hDrive != INVALID_HANDLE_VALUE) {
        CloseHandle(hDrive);
    }
    if (hISO != INVALID_HANDLE_VALUE) {
        CloseHandle(hISO);
    }

    return success ? 0 : -1;
}

/*
 * Extract disk number from device path.
 * Handles formats like "\\.\PhysicalDrive2", "PhysicalDrive2", or plain "2".
 * Returns -1 on invalid input.
 */
static int extract_disk_number(const char *device) {
    if (device == NULL || device[0] == '\0') {
        return -1;
    }

    /* Skip \\.\  prefix if present */
    const char *p = device;
    if (strncmp(p, "\\\\.\\", 4) == 0) {
        p += 4;
    }

    /* Skip "PhysicalDrive" prefix if present (case-insensitive) */
    if (_strnicmp(p, "PhysicalDrive", 13) == 0) {
        p += 13;
    }

    /* Now p should point to the disk number */
    if (*p == '\0') {
        return -1;
    }

    /* Verify remaining characters are all digits */
    const char *start = p;
    while (*p != '\0') {
        if (!isdigit((unsigned char)*p)) {
            return -1;
        }
        p++;
    }

    return atoi(start);
}

/*
 * Case-insensitive string comparison for filesystem validation.
 * Returns 1 if strings match (ignoring case), 0 otherwise.
 */
static int strcasecmp_fs(const char *a, const char *b) {
    return _stricmp(a, b) == 0;
}

/*
 * Run diskpart with a script file.
 * Returns exit code (0 = success), or -1 on execution failure.
 */
static int run_diskpart(const char *script_path) {
    STARTUPINFOA si;
    PROCESS_INFORMATION pi;
    char cmd_line[512];
    DWORD exit_code;
    DWORD wait_result;

    /* Build command line: diskpart /s "scriptpath" */
    snprintf(cmd_line, sizeof(cmd_line), "diskpart.exe /s \"%s\"", script_path);

    ZeroMemory(&si, sizeof(si));
    si.cb = sizeof(si);
    si.dwFlags = STARTF_USESHOWWINDOW;
    si.wShowWindow = SW_HIDE;  /* Hide console window */
    ZeroMemory(&pi, sizeof(pi));

    /* Create the diskpart process */
    if (!CreateProcessA(
            NULL,           /* Use cmd_line for executable */
            cmd_line,       /* Command line */
            NULL,           /* Process security attributes */
            NULL,           /* Thread security attributes */
            FALSE,          /* Don't inherit handles */
            CREATE_NO_WINDOW, /* No console window */
            NULL,           /* Use parent's environment */
            NULL,           /* Use parent's working directory */
            &si,
            &pi)) {
        set_last_error_message("failed to execute diskpart: error %lu", GetLastError());
        return -1;
    }

    /* Wait for diskpart to complete (5 minute timeout for slow USB drives) */
    wait_result = WaitForSingleObject(pi.hProcess, 5 * 60 * 1000);

    if (wait_result == WAIT_TIMEOUT) {
        TerminateProcess(pi.hProcess, 1);
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);
        set_last_error_message("diskpart timed out after 5 minutes");
        return -1;
    }

    if (wait_result != WAIT_OBJECT_0) {
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);
        set_last_error_message("WaitForSingleObject failed: error %lu", GetLastError());
        return -1;
    }

    /* Get the exit code */
    if (!GetExitCodeProcess(pi.hProcess, &exit_code)) {
        exit_code = 1;  /* Assume failure if we can't get exit code */
    }

    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);

    return (int)exit_code;
}

int format_disk(const char *device, const char *filesystem, const char *volume_name) {
    char temp_path[MAX_PATH];
    char script_path[MAX_PATH];
    FILE *script_file;
    int disk_num;
    int result;
    const char *fs_upper;

    /* Validate device path */
    if (device == NULL || device[0] == '\0') {
        set_last_error_message("device path is required");
        return -1;
    }

    /* Validate filesystem type */
    if (filesystem == NULL || filesystem[0] == '\0') {
        set_last_error_message("filesystem type is required");
        return -1;
    }

    /* Use empty string if volume_name is NULL */
    if (volume_name == NULL) {
        volume_name = "";
    }

    /* Reset cancellation flag before starting */
    reset_cancel_flag();

    /* Extract and validate disk number */
    disk_num = extract_disk_number(device);
    if (disk_num < 0) {
        set_last_error_message("invalid device path: %s (expected PhysicalDriveN format)", device);
        return -1;
    }

    /* Safety check: refuse to format disk 0 (system disk) */
    if (disk_num == 0) {
        set_last_error_message("refusing to format disk 0 (system disk)");
        return -1;
    }

    /* Validate and normalize filesystem type */
    if (strcasecmp_fs(filesystem, "FAT32")) {
        fs_upper = "FAT32";
    } else if (strcasecmp_fs(filesystem, "NTFS")) {
        fs_upper = "NTFS";
    } else if (strcasecmp_fs(filesystem, "exFAT")) {
        fs_upper = "exFAT";
    } else {
        set_last_error_message("unsupported filesystem: %s (use FAT32, NTFS, or exFAT)", filesystem);
        return -1;
    }

    /* Get temp directory */
    if (GetTempPathA(MAX_PATH, temp_path) == 0) {
        set_last_error_message("failed to get temp path: error %lu", GetLastError());
        return -1;
    }

    /* Create temp script file */
    if (GetTempFileNameA(temp_path, "fid", 0, script_path) == 0) {
        set_last_error_message("failed to create temp file: error %lu", GetLastError());
        return -1;
    }

    /* Write diskpart script */
    script_file = fopen(script_path, "w");
    if (script_file == NULL) {
        DeleteFileA(script_path);
        set_last_error_message("failed to open temp script file for writing");
        return -1;
    }

    /* Write diskpart commands */
    fprintf(script_file, "select disk %d\n", disk_num);
    fprintf(script_file, "clean\n");
    fprintf(script_file, "create partition primary\n");

    /* Format command with optional label */
    if (volume_name[0] != '\0') {
        fprintf(script_file, "format fs=%s quick label=\"%s\"\n", fs_upper, volume_name);
    } else {
        fprintf(script_file, "format fs=%s quick\n", fs_upper);
    }

    fprintf(script_file, "active\n");
    fprintf(script_file, "exit\n");
    fclose(script_file);

    /* Check for cancellation before running diskpart */
    if (is_cancelled()) {
        DeleteFileA(script_path);
        set_last_error_message("operation cancelled");
        return -1;
    }

    /* Run diskpart with the script */
    result = run_diskpart(script_path);

    /* Clean up temp file */
    DeleteFileA(script_path);

    /* Check result */
    if (result != 0) {
        if (result > 0) {
            set_last_error_message("diskpart failed with exit code %d", result);
        }
        /* If result == -1, error message already set by run_diskpart */
        return -1;
    }

    return 0;
}
