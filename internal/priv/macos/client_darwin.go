//go:build darwin && cgo && osinstallxpc

package macos

/*
#cgo darwin CFLAGS: -x objective-c -fmodules -fobjc-arc
#cgo darwin LDFLAGS: -framework Foundation -framework ServiceManagement -framework Security -framework CoreFoundation

#include <stdlib.h>

// C-callable functions implemented in xpc_client.m
int helper_ensure_ready(char **errmsg);
int helper_unmount_disk(const char *device, char **out, char **errmsg);
int helper_eject_disk(const char *device, char **out, char **errmsg);
int helper_write_linux_iso(const char *device, const char *isoPath, char **errmsg);
*/
import "C"

import (
	"context"
	"errors"
	"unsafe"
)

type xpcClient struct{}

func NewClient() Client { return &xpcClient{} }

func (c *xpcClient) EnsureReady(ctx context.Context) error {
	var cerr *C.char
	ret := C.helper_ensure_ready(&cerr)
	if ret != 0 {
		defer C.free(unsafe.Pointer(cerr))
		return errors.New(C.GoString(cerr))
	}
	return nil
}

func (c *xpcClient) UnmountDisk(ctx context.Context, device string) (string, error) {
	cd := C.CString(device)
	defer C.free(unsafe.Pointer(cd))
	var cout *C.char
	var cerr *C.char
	ret := C.helper_unmount_disk(cd, &cout, &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
		}
		return "", errors.New(C.GoString(cerr))
	}
	defer C.free(unsafe.Pointer(cout))
	return C.GoString(cout), nil
}

func (c *xpcClient) EjectDisk(ctx context.Context, device string) (string, error) {
	cd := C.CString(device)
	defer C.free(unsafe.Pointer(cd))
	var cout *C.char
	var cerr *C.char
	ret := C.helper_eject_disk(cd, &cout, &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
		}
		return "", errors.New(C.GoString(cerr))
	}
	defer C.free(unsafe.Pointer(cout))
	return C.GoString(cout), nil
}

// WriteLinuxISO writes a Linux ISO to a disk using Disk Arbitration and direct I/O.
// This is a long-running operation that can take several minutes for large ISOs.
// The pipeline is: DA session → claim disk → unmount → raw write → eject → unclaim
func (c *xpcClient) WriteLinuxISO(ctx context.Context, isoPath string, device string) error {
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
