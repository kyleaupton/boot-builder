//go:build (darwin || linux) && wimlib

package wim

/*
#include <wimlib.h>
#include <stdlib.h>
int register_progress(WIMStruct *w, uintptr_t ctx);
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"runtime/cgo"
	"unsafe"
)

type splitContext struct {
	ctx context.Context
	cb  func(Progress) bool
}

func SplitWithProgress(ctx context.Context, src, dstPrefix string, partSizeMiB int, check bool, cb func(Progress) bool) error {
	if err := globalInit(); err != nil {
		return err
	}
	if partSizeMiB <= 0 {
		return errors.New("partSizeMiB must be > 0")
	}

	cSrc := C.CString(src)
	cDst := C.CString(dstPrefix)
	defer func() {
		C.free(unsafe.Pointer(cSrc))
		C.free(unsafe.Pointer(cDst))
	}()

	var wim *C.WIMStruct
	var openFlags C.int = 0
	if check {
		openFlags |= C.WIMLIB_OPEN_FLAG_CHECK_INTEGRITY
	}
	if rc := C.wimlib_open_wim(cSrc, openFlags, &wim); rc != 0 {
		return fmt.Errorf("wimlib_open_wim failed: code %d", int(rc))
	}
	defer C.wimlib_free(wim)

	// Register progress callback using a handle that stores ctx+cb.
	state := &splitContext{ctx: ctx, cb: cb}
	h := cgo.NewHandle(state)
	defer h.Delete()
	if rc := C.register_progress(wim, (C.uintptr_t)(uintptr(h))); rc != 0 {
		return fmt.Errorf("register_progress failed: code %d", int(rc))
	}

	partBytes := C.uint64_t(uint64(partSizeMiB) * 1024 * 1024)
	var splitFlags C.int = 0
	rc := C.wimlib_split(wim, cDst, partBytes, splitFlags)
	if rc != 0 {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}
		return fmt.Errorf("wimlib_split failed: code %d", int(rc))
	}
	return nil
}
