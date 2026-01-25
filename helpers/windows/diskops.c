#include "diskops.h"
#include "pipe.h"
#include <windows.h>
#include <stdio.h>
#include <stdarg.h>
#include <string.h>
#include <ctype.h>

/*
 * Disk operations implementation for FlashIt privileged helper.
 * format_disk: Implemented using diskpart
 * write_iso: STUB - pending implementation
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

int send_progress_update(OperationContext *ctx, uint64_t written, uint64_t total) {
    if (ctx == NULL || ctx->pipe == NULL || ctx->pipe == INVALID_HANDLE_VALUE) {
        return -1;
    }

    Response resp;
    make_progress_response(&resp, ctx->request_id, written, total);
    return send_response(ctx->pipe, &resp);
}

int write_iso(const char *device, const char *iso_path, OperationContext *ctx) {
    /*
     * STUB IMPLEMENTATION
     *
     * TODO: Implement the following steps:
     * 1. Validate device path (must be PhysicalDrive format)
     * 2. Open the ISO file for reading
     * 3. Get ISO file size for progress reporting
     * 4. Open the physical drive for writing (CreateFile with GENERIC_WRITE)
     * 5. Lock the volume (DeviceIoControl FSCTL_LOCK_VOLUME)
     * 6. Dismount the volume (DeviceIoControl FSCTL_DISMOUNT_VOLUME)
     * 7. Read ISO in chunks, write to device
     * 8. Send progress updates periodically
     * 9. Check cancellation flag between chunks
     * 10. Unlock volume and close handles
     * 11. Flush device buffers
     */

    (void)ctx;  /* Suppress unused parameter warning */

    if (device == NULL || device[0] == '\0') {
        set_last_error_message("device path is required");
        return -1;
    }

    if (iso_path == NULL || iso_path[0] == '\0') {
        set_last_error_message("ISO path is required");
        return -1;
    }

    /* Reset cancellation flag before starting */
    reset_cancel_flag();

    set_last_error_message("write_iso: not implemented - raw disk write pending");
    return -1;
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
