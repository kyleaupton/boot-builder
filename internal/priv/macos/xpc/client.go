//go:build darwin && cgo && osinstallxpc

package xpc

/*
#cgo darwin CFLAGS: -x objective-c -fmodules -fobjc-arc
#cgo darwin LDFLAGS: -framework Foundation -framework ServiceManagement -framework Security -framework CoreFoundation

#include <stdlib.h>
#include <stdint.h>

// C-callable functions implemented in xpc_client.m
int helper_ensure_ready(char **errmsg);
int helper_write_linux_iso(const char *device, const char *isoPath, char **errmsg);
int helper_format_disk(const char *device, const char *filesystem, const char *volumeName, char **errmsg);
void helper_cancel_current_operation(void);
*/
import "C"

import (
	"context"
	"errors"
	"sync"
	"unsafe"
)

// ProgressFunc is the callback type for write progress updates.
type ProgressFunc func(bytesWritten, totalBytes uint64)

// progressCallback stores the current progress callback.
// Only one write operation is supported at a time.
var (
	progressCallback ProgressFunc
	progressMu       sync.Mutex
)

// GoProgressCallback is called from Objective-C when the helper reports progress.
// This function is exported to C via CGO.
//
//export GoProgressCallback
func GoProgressCallback(written, total C.uint64_t) {
	progressMu.Lock()
	cb := progressCallback
	progressMu.Unlock()
	if cb != nil {
		cb(uint64(written), uint64(total))
	}
}

// Client provides XPC communication with the privileged helper.
type Client struct{}

// NewClient creates a new XPC client.
func NewClient() *Client { return &Client{} }

func (c *Client) EnsureReady(ctx context.Context) error {
	var cerr *C.char
	ret := C.helper_ensure_ready(&cerr)
	if ret != 0 {
		defer C.free(unsafe.Pointer(cerr))
		return errors.New(C.GoString(cerr))
	}
	return nil
}

// ErrCancelled is returned when an operation is cancelled by the user.
var ErrCancelled = errors.New("operation cancelled")

// CancelCurrentOperation cancels the currently running operation (if any).
// This is safe to call even if no operation is running.
func (c *Client) CancelCurrentOperation() {
	C.helper_cancel_current_operation()
}

// WriteLinuxISO writes a Linux ISO to a disk using Disk Arbitration and direct I/O.
// This is a long-running operation that can take several minutes for large ISOs.
// The pipeline is: DA session → claim disk → unmount → raw write → eject → unclaim
//
// The progress callback receives (bytesWritten, totalBytes) updates during the write.
// Pass nil if progress updates are not needed.
//
// If the context is cancelled, the operation will be cancelled and ErrCancelled returned.
func (c *Client) WriteLinuxISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	// Set the progress callback (only one write at a time is supported)
	progressMu.Lock()
	progressCallback = progress
	progressMu.Unlock()
	defer func() {
		progressMu.Lock()
		progressCallback = nil
		progressMu.Unlock()
	}()

	// Set up context cancellation goroutine
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			c.CancelCurrentOperation()
		case <-done:
			// Operation completed normally, exit goroutine
		}
	}()

	cdev := C.CString(device)
	defer C.free(unsafe.Pointer(cdev))
	cpath := C.CString(isoPath)
	defer C.free(unsafe.Pointer(cpath))
	var cerr *C.char

	ret := C.helper_write_linux_iso(cdev, cpath, &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
			errMsg := C.GoString(cerr)
			// Check for cancellation error (prefixed with "cancelled:")
			if len(errMsg) > 10 && errMsg[:10] == "cancelled:" {
				return ErrCancelled
			}
			return errors.New(errMsg)
		}
		return errors.New("failed to write Linux ISO to disk")
	}

	return nil
}

// FormatDisk formats a disk with the specified filesystem and volume name.
// This uses diskutil eraseDisk under the hood.
func (c *Client) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	cdev := C.CString(device)
	defer C.free(unsafe.Pointer(cdev))
	cfs := C.CString(filesystem)
	defer C.free(unsafe.Pointer(cfs))
	cname := C.CString(volumeName)
	defer C.free(unsafe.Pointer(cname))
	var cerr *C.char

	ret := C.helper_format_disk(cdev, cfs, cname, &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
			return errors.New(C.GoString(cerr))
		}
		return errors.New("failed to format disk")
	}

	return nil
}
