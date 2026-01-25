# CLAUDE.md

This file provides guidance for Claude Code when working on this project.

Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask, except for things that have explicit MCP servers like shadcn-vue.

Documentation IDs:
Wails3 - /websites/v3alpha_wails_io

## Project Overview

**FlashIt** is a cross-platform desktop app built with Wails v3 that creates bootable USB OS installers. The primary motivation: making Windows USB installers on macOS requires splitting large `.wim` files (>4GB) for FAT32 compatibility using wimlib.

**Branch:** `go-rewrite` (active development branch)

## Tech Stack

- **Backend:** Go 1.21+ with Wails v3-alpha
- **Frontend:** Vue 3 + TypeScript + Vite
- **Build:** Wails Task framework (`task dev` for development)
- **CGO:** wimlib integration for Windows media splitting

## Project Structure

```
/
├── main.go                    # Wails v3 entry point
├── internal/
│   ├── core/                  # Type contracts (Plan, Step, Installer, Event)
│   ├── service/               # Wails services (JobsService, DrivesService)
│   ├── jobs/                  # Job execution engine
│   ├── installers/            # OS-specific installers
│   │   └── linux/ubuntu/      # Ubuntu installer (working on macOS)
│   ├── steps/                 # Executable job steps (disk ops)
│   ├── drives/                # Cross-platform drive detection
│   ├── priv/                  # Privileged command execution
│   ├── fs/                    # Low-level file operations
│   ├── wim/                   # CGO bridge to wimlib
│   └── eventbus/              # Global event emitter
├── frontend/
│   ├── src/                   # Vue components
│   └── bindings/              # Auto-generated Wails bindings
├── build/                     # Platform-specific build configs
├── macos/privilaged/          # macOS privileged helper daemon
└── docs/                      # Architecture documentation
```

## Key Commands

```bash
# Development (hot-reload)
task dev

# Build for current platform
task build

# Generate frontend bindings (auto-runs on dev)
wails3 generate bindings
```

## Architecture Patterns

### Installer Interface
Each OS installer implements `core.Installer`:
```go
type Installer interface {
    ID() string
    Name() string
    Targets() []Target
    ValidateHost(ctx context.Context) bool
    Plan(ctx context.Context, target Target, sourcePath string, drive drives.Drive) (*Plan, error)
}
```

### Job Execution Flow
1. `JobsService.StartJob()` creates a `Plan` with `Steps`
2. `jobs.Manager.Enqueue()` runs steps sequentially
3. Steps emit `core.Event` via `eventbus.Emit()`
4. Wails forwards events to frontend

### Privilege Handling
- **macOS:** Launchd daemon + XPC socket (`macos/privilaged/`)
- **Linux:** pkexec (stubbed)
- **Windows:** UAC (stubbed)

## Current Implementation Status

### Working (macOS)
- Drive detection (`diskutil` parsing)
- Disk operations (unmount, raw write, eject)
- Ubuntu ISO → USB flow
- wimlib CGO bridge (splitting ready, not integrated)

### Stubbed (needs implementation)
- Linux drive detection (`/sys` + udev)
- Windows drive detection (WMI)
- Linux/Windows privilege elevation
- Windows installer (wimlib integration)

## WIM File Handling

Located in `internal/wim/`. Pure Go implementation for reading and splitting Windows `.wim` files for FAT32 compatibility (files >4GB must be split).

```go
// Split a WIM file for FAT32 compatibility
wim.SplitWithProgress(ctx, srcWIM, dstPrefix, opts, progressFn)

// Copy split SWM files to USB
wim.CopySWMs(ctx, swmDir, usbRoot, progressFn)
```

## Event Schema

Events emitted to frontend via `job:event`:
```go
type Event struct {
    JobID   string
    Type    string   // "state", "step-start", "step-end", "progress", "log", "error"
    Message string
    Step    string
    Percent int
    Error   string
}
```

## Important Notes

- Never write to `/dev/disk0` (boot drive protection in `internal/installers/linux/ubuntu/ubuntu.go`)
- Raw device writes use `/dev/rdisk*` on macOS for speed
- FAT32 max file size is ~4GB, hence wimlib splitting for Windows
- Privileged helper verifies client code signature before accepting commands

## Documentation

See `/docs/` for detailed architecture docs:
- `ARCHITECTURE.md` - High-level overview
- `BACKEND.md` - Services, jobs, steps
- `MODULES/*.md` - Per-module deep dives
