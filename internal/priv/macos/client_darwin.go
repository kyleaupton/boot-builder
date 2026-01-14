//go:build darwin && cgo && osinstallxpc

package macos

/*
#cgo darwin CFLAGS: -x objective-c -fmodules -fobjc-arc
#cgo darwin LDFLAGS: -framework Foundation -framework ServiceManagement -framework Security -framework CoreFoundation

#include <stdlib.h>

typedef void (*progress_cb_t)(long long wrote, long long total);

// C-callable functions implemented in xpc_client.m
int helper_ensure_ready(char **errmsg);
int helper_unmount_disk(const char *device, char **out, char **errmsg);
int helper_eject_disk(const char *device, char **out, char **errmsg);
int helper_raw_write(const char *iso_path, const char *raw_device, progress_cb_t cb, char **errmsg);

// Trampoline to call Go progress callback from C
extern void go_progress_trampoline(long long wrote, long long total);
static void progress_adapter(progress_cb_t cb, long long wrote, long long total) { cb(wrote, total); }
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

//export go_progress_trampoline
func go_progress_trampoline(wrote C.longlong, total C.longlong) {}

func (c *xpcClient) RawWrite(ctx context.Context, isoPath string, rawDevice string, onProgress func(wrote, total int64)) error {
	ciso := C.CString(isoPath)
	cdev := C.CString(rawDevice)
	defer C.free(unsafe.Pointer(ciso))
	defer C.free(unsafe.Pointer(cdev))
	var cerr *C.char
	// For now, no progress bridge. We can wire it later using a global callback.
	ret := C.helper_raw_write(ciso, cdev, (C.progress_cb_t)(C.go_progress_trampoline), &cerr)
	if ret != 0 {
		if cerr != nil {
			defer C.free(unsafe.Pointer(cerr))
		}
		return errors.New(C.GoString(cerr))
	}
	return nil
}
