//go:build (darwin || linux) && wimlib

package wim

/*
#cgo CFLAGS: -I${SRCDIR}/native/include
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/native/lib/darwin_arm64 -lwim
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/native/lib/darwin_amd64 -lwim
#cgo linux,amd64  LDFLAGS: -L${SRCDIR}/native/lib/linux_amd64  -lwim
#include <wimlib.h>
#include <stdint.h>
#include <stdlib.h>
#include "adapter.h"

extern int goProgressTrampoline(int msg, struct go_progress_info *gi, void *ctx);

static int progress_adapter(enum wimlib_progress_msg msg, union wimlib_progress_info *info, void *ctx) {
    struct go_progress_info gi;
    fill_go_progress_info(msg, info, &gi);
    return goProgressTrampoline((int)msg, &gi, ctx);
}

static int register_progress(WIMStruct *w, uintptr_t ctx) {
    return wimlib_register_progress_function(w, progress_adapter, (void*)ctx);
}
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"
)

var initOnce sync.Once
var initErr error

func globalInit() error {
	initOnce.Do(func() {
		if rc := C.wimlib_global_init(); rc != 0 {
			initErr = fmt.Errorf("wimlib_global_init failed: code %d", int(rc))
		}
	})
	return initErr
}

//export goProgressTrampoline
func goProgressTrampoline(msg C.int, gi *C.struct_go_progress_info, ctx unsafe.Pointer) C.int {
	h := cgo.Handle(uintptr(ctx))
	state, ok := h.Value().(*splitContext)
	if !ok || state == nil {
		return C.WIMLIB_PROGRESS_STATUS_CONTINUE
	}
	// Cancellation first
	select {
	case <-state.ctx.Done():
		return C.WIMLIB_PROGRESS_STATUS_ABORT
	default:
	}

	p := Progress{
		Phase:          mapPhase(int(msg), int(gi.phase_hint)),
		BytesCompleted: uint64(gi.bytes_completed),
		BytesTotal:     uint64(gi.bytes_total),
		ItemsCompleted: uint32(gi.items_completed),
		ItemsTotal:     uint32(gi.items_total),
		PartIndex:      uint32(gi.part_index),
		PartCount:      uint32(gi.part_count),
	}
	if p.BytesTotal > 0 {
		p.Percent = float64(p.BytesCompleted) * 100.0 / float64(p.BytesTotal)
	} else if p.ItemsTotal > 0 {
		p.Percent = float64(p.ItemsCompleted) * 100.0 / float64(p.ItemsTotal)
	}

	if state.cb != nil && !state.cb(p) {
		return C.WIMLIB_PROGRESS_STATUS_ABORT
	}
	return C.WIMLIB_PROGRESS_STATUS_CONTINUE
}
