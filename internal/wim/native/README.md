This directory holds the vendored libwim headers and shared libraries for runtime linking.

Expected layout:

native/
  include/
    wimlib.h
  lib/
    darwin_arm64/
      libwim.dylib
    darwin_amd64/
      libwim.dylib
    linux_amd64/
      libwim.so

Notes:
- macOS: ensure install_name is `@rpath/libwim.dylib` and our LDFLAGS set an rpath to this folder.
- Linux: we set rpath to `$ORIGIN/../native/lib/linux_amd64` so the binary can find `libwim.so` at runtime.
- You may generate these artifacts in CI or fetch from a release during development.
- Keep versions in sync across OS/arch.


