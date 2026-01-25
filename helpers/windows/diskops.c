#include "diskops.h"
#include "pipe.h"
#include <windows.h>
#include <stdio.h>
#include <stdarg.h>

/*
 * Disk operations implementation for FlashIt privileged helper.
 * STUB IMPLEMENTATION - Returns errors for all operations.
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

int format_disk(const char *device, const char *filesystem, const char *volume_name) {
    /*
     * STUB IMPLEMENTATION
     *
     * TODO: Implement the following steps:
     * 1. Validate device path
     * 2. Validate filesystem type (FAT32, NTFS, exFAT)
     * 3. Use Windows Format API or shell out to format.com
     * 4. Alternative: Use DeviceIoControl for low-level formatting
     * 5. Set volume label
     */

    if (device == NULL || device[0] == '\0') {
        set_last_error_message("device path is required");
        return -1;
    }

    if (filesystem == NULL || filesystem[0] == '\0') {
        set_last_error_message("filesystem type is required");
        return -1;
    }

    if (volume_name == NULL) {
        volume_name = "";
    }

    /* Reset cancellation flag before starting */
    reset_cancel_flag();

    set_last_error_message("format_disk: not implemented - disk formatting pending");
    return -1;
}
