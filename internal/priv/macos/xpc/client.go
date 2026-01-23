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

// WriteLinuxISO writes a Linux ISO to a disk using Disk Arbitration and direct I/O.
// This is a long-running operation that can take several minutes for large ISOs.
// The pipeline is: DA session → claim disk → unmount → raw write → eject → unclaim
//
// The progress callback receives (bytesWritten, totalBytes) updates during the write.
// Pass nil if progress updates are not needed.
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

	cdev := C.CString(device)
	defer C.free(unsafe.Pointer(cdev))
	cpath := C.CString(isoPath)
	defer C.free(unsafe.Pointer(cpath))
	var cerr *C.char

	ret := C.helper_write_linux_iso(cdev, cpath, &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
			return errors.New(C.GoString(cerr))
		}
		return errors.New("failed to write Linux ISO to disk")
	}

	return nil
}
