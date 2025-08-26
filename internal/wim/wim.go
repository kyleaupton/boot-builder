package wim

/*
#cgo darwin,arm64 CFLAGS:  -I${SRCDIR}/deps/darwin-arm64/include
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/deps/darwin-arm64/lib -lwim

#include <stdlib.h>
#include <stdint.h>
#include <wimlib.h>

// The bridge is defined in bridge.c (which includes _cgo_export.h)
int progress_trampoline_bridge(int msg,
                               const union wimlib_progress_info* info,
                               void* user);

// Weak-import the optional APIs, as before (trim if you already have these in place)
#if defined(__APPLE__)
__attribute__((weak_import)) int  wimlib_open_wim_with_progress(const wimlib_tchar *wim_file,
                                                                int open_flags,
                                                                WIMStruct **wim_ret,
                                                                wimlib_progress_func_t progfunc,
                                                                void *progctx);
__attribute__((weak_import)) void wimlib_register_progress_function(WIMStruct *w,
                                                                    wimlib_progress_func_t progfunc,
                                                                    void *progctx);
#else
int  wimlib_open_wim_with_progress(const wimlib_tchar *wim_file,
                                   int open_flags,
                                   WIMStruct **wim_ret,
                                   wimlib_progress_func_t progfunc,
                                   void *progctx);
void wimlib_register_progress_function(WIMStruct *w,
                                       wimlib_progress_func_t progfunc,
                                       void *progctx);
#endif

// Our C-side progress function → calls the bridge (which calls the Go export)
static int progress_trampoline(enum wimlib_progress_msg msg,
                               const union wimlib_progress_info* info,
                               void* user) {
    return progress_trampoline_bridge((int)msg, info, user);
}

static int wim_split_with_cb(const char* src,
                             const char* dst_prefix,
                             uint64_t part_size_bytes,
                             int check_integrity,
                             void* user) {
    int ret;
    WIMStruct* w = NULL;

    int open_flags = 0;
#ifdef WIMLIB_OPEN_FLAG_CHECK_INTEGRITY
    if (check_integrity) {
        open_flags |= WIMLIB_OPEN_FLAG_CHECK_INTEGRITY;
    }
#endif

    if (wimlib_open_wim_with_progress) {
        // NOTE: your header’s order: (path, flags, **wim_ret**, progfunc, progctx)
        ret = wimlib_open_wim_with_progress((const wimlib_tchar*)src,
                                            open_flags,
                                            &w,
                                            (wimlib_progress_func_t)progress_trampoline,
                                            user);
        if (ret != 0) return ret;
    } else {
        ret = wimlib_open_wim((const wimlib_tchar*)src, open_flags, &w);
        if (ret != 0) return ret;
        if (wimlib_register_progress_function) {
            wimlib_register_progress_function(w,
                    (wimlib_progress_func_t)progress_trampoline, user);
        }
    }

    // 4-arg split: pass 0 for write_flags per your header
    ret = wimlib_split(w, dst_prefix, part_size_bytes, 0);

    wimlib_free(w);
    return ret;
}
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"runtime/cgo"
	"unsafe"
)

type Progress struct {
	Phase      string
	DoneBytes  uint64
	TotalBytes uint64
	Part       uint32
	TotalParts uint32
}

type SplitOptions struct {
	PartSizeMiB    int
	CheckIntegrity bool
}

var libInited bool

func Init() error {
	if libInited {
		return nil
	}
	if ret := C.wimlib_global_init(C.int(0)); ret != 0 {
		return fmt.Errorf("wimlib_global_init failed: %d", int(ret))
	}
	libInited = true
	return nil
}

func SplitWithProgress(
	ctx context.Context,
	srcWIM string,
	dstPrefix string,
	opts SplitOptions,
	cb func(Progress) bool,
) error {
	if err := Init(); err != nil {
		return err
	}
	if opts.PartSizeMiB <= 0 {
		opts.PartSizeMiB = 3800
	}

	cSrc := C.CString(srcWIM)
	defer C.free(unsafe.Pointer(cSrc))
	cDst := C.CString(dstPrefix)
	defer C.free(unsafe.Pointer(cDst))

	h := cgo.NewHandle(progressState{ctx: ctx, cb: cb})
	defer h.Delete()

	partBytes := C.uint64_t(uint64(opts.PartSizeMiB) * 1024 * 1024)
	check := C.int(0)
	if opts.CheckIntegrity {
		check = 1
	}

	ret := C.wim_split_with_cb(cSrc, cDst, partBytes, check, unsafe.Pointer(h))
	if ret == 0 {
		return nil
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return context.Canceled
	}
	return fmt.Errorf("wimlib split failed: %d", int(ret))
}

type progressState struct {
	ctx context.Context
	cb  func(Progress) bool
}

//export goWimProgress
func goWimProgress(msg C.int, info *C.union_wimlib_progress_info, user unsafe.Pointer) C.int {
	h := cgo.Handle(user)
	st, ok := h.Value().(progressState)
	if !ok {
		return 0
	}
	select {
	case <-st.ctx.Done():
		return 1
	default:
	}

	var p Progress
	switch msg {
	case C.WIMLIB_PROGRESS_MSG_WRITE_STREAMS:
		ws := (*C.struct_wimlib_progress_info_write_streams)(unsafe.Pointer(info))
		p.Phase = "writing"
		p.DoneBytes = uint64(ws.completed_bytes)
		p.TotalBytes = uint64(ws.total_bytes)

	case C.WIMLIB_PROGRESS_MSG_SPLIT_BEGIN_PART:
		sp := (*C.struct_wimlib_progress_info_split)(unsafe.Pointer(info))
		p.Phase = "split-begin-part"
		p.Part = uint32(sp.cur_part_number)
		p.TotalParts = uint32(sp.total_parts)
		p.DoneBytes = uint64(sp.completed_bytes) // overall bytes copied so far
		p.TotalBytes = uint64(sp.total_bytes)    // overall bytes to copy

	case C.WIMLIB_PROGRESS_MSG_SPLIT_END_PART:
		sp := (*C.struct_wimlib_progress_info_split)(unsafe.Pointer(info))
		p.Phase = "split-end-part"
		p.Part = uint32(sp.cur_part_number)
		p.TotalParts = uint32(sp.total_parts)
		p.DoneBytes = uint64(sp.completed_bytes)
		p.TotalBytes = uint64(sp.total_bytes)

	// If your header exposes integrity progress structs and you opened with CHECK_INTEGRITY,
	// you can also add:
	// case C.WIMLIB_PROGRESS_MSG_VERIFY_INTEGRITY:
	//   vi := (*C.struct_wimlib_progress_info_integrity)(unsafe.Pointer(info))
	//   p.Phase = "verify"
	//   p.DoneBytes = uint64(vi.completed_bytes)
	//   p.TotalBytes = uint64(vi.total_bytes)

	default:
		p.Phase = "other"
	}
	if st.cb != nil && !st.cb(p) {
		return 1
	}
	return 0
}
