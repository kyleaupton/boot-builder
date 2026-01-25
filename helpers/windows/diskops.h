#ifndef FLASHIT_DISKOPS_H
#define FLASHIT_DISKOPS_H

#include <stdint.h>
#include "protocol.h"

/*
 * Disk operations for FlashIt privileged helper.
 * These functions perform raw disk access and require administrator privileges.
 */

/* Forward declaration of pipe handle for sending progress updates */
typedef void* HANDLE;

/* Progress callback function type */
typedef void (*progress_callback_t)(uint64_t written, uint64_t total, void *user_data);

/* Context for progress reporting during operations */
typedef struct {
    const char *request_id;
    HANDLE pipe;
    volatile int *cancel_flag;
} OperationContext;

/*
 * Write an ISO image to a raw disk device.
 *
 * device: Physical device path (e.g., "\\\\.\\PhysicalDrive2")
 * iso_path: Path to the ISO file to write
 * ctx: Operation context for progress reporting and cancellation
 *
 * Returns 0 on success, -1 on error.
 * On error, use get_last_error_message() to retrieve the error description.
 */
int write_iso(const char *device, const char *iso_path, OperationContext *ctx);

/*
 * Format a disk with the specified filesystem.
 *
 * device: Physical device path (e.g., "\\\\.\\PhysicalDrive2")
 * filesystem: Filesystem type ("FAT32", "NTFS", "exFAT")
 * volume_name: Volume label for the new filesystem
 *
 * Returns 0 on success, -1 on error.
 * On error, use get_last_error_message() to retrieve the error description.
 */
int format_disk(const char *device, const char *filesystem, const char *volume_name);

/*
 * Cancel the currently running operation.
 * This sets a flag that the write_iso/format_disk functions check periodically.
 * Safe to call from any thread.
 */
void cancel_current_operation(void);

/*
 * Reset the cancellation flag.
 * Call this before starting a new operation.
 */
void reset_cancel_flag(void);

/*
 * Check if cancellation was requested.
 * Returns non-zero if cancelled.
 */
int is_cancelled(void);

/*
 * Get the last error message from disk operations.
 * Returns a pointer to a thread-local static buffer.
 */
const char* get_last_error_message(void);

/*
 * Set the last error message.
 * This is used internally by disk operations.
 */
void set_last_error_message(const char *fmt, ...);

/*
 * Send a progress update over the pipe.
 * Returns 0 on success, -1 on error.
 */
int send_progress_update(OperationContext *ctx, uint64_t written, uint64_t total);

#endif /* FLASHIT_DISKOPS_H */
