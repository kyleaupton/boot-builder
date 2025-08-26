//go:build (darwin || linux) && wimlib

#pragma once
#include <stdint.h>

// Flattened progress payload passed from C to Go.
struct go_progress_info {
    int msg;                  // enum wimlib_progress_msg
    uint64_t bytes_completed;
    uint64_t bytes_total;
    uint32_t items_completed;
    uint32_t items_total;
    uint32_t part_index;      // 1-based if available
    uint32_t part_count;
    int      phase_hint;      // optional mapping aid
};

// Populates go_progress_info from wimlib's union wimlib_progress_info.
// The implementation may leave fields zero when not applicable.
void fill_go_progress_info(int msg, void *info_union, struct go_progress_info *out);


