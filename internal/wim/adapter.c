//go:build (darwin || linux) && wimlib

#include <wimlib.h>
#include "adapter.h"

void fill_go_progress_info(int msg, void *info_union, struct go_progress_info *out) {
    union wimlib_progress_info *u = (union wimlib_progress_info*)info_union;
    out->msg = msg;
    out->bytes_completed = 0;
    out->bytes_total = 0;
    out->items_completed = 0;
    out->items_total = 0;
    out->part_index = 0;
    out->part_count = 0;
    out->phase_hint = 0;

    switch (msg) {
        case WIMLIB_PROGRESS_MSG_WRITE_STREAMS:
            out->bytes_completed = u->write_streams.completed_bytes;
            out->bytes_total     = u->write_streams.total_bytes;
            out->items_completed = u->write_streams.completed_streams;
            out->items_total     = u->write_streams.total_streams;
            out->phase_hint      = 2; // writing streams
            break;
        case WIMLIB_PROGRESS_MSG_VERIFY_INTEGRITY:
            out->bytes_completed = u->verify_integrity.completed_bytes;
            out->bytes_total     = u->verify_integrity.total_bytes;
            out->phase_hint      = 4; // verifying
            break;
        default:
            break;
    }
}


