
#include "_cgo_export.h"
#include <wimlib.h>

// Use a separate symbol we can reference from the preamble.
int progress_trampoline_bridge(int msg,
                               const union wimlib_progress_info* info,
                               void* user) {
  // Match the exact prototype cgo generated for goWimProgress by including _cgo_export.h
  // We drop const on info when passing to Go (we only read).
  return (int)goWimProgress((int)msg, (union wimlib_progress_info*)info, user);
}