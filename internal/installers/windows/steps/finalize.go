//go:build darwin

package steps

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"boot-builder/internal/core"
	"boot-builder/internal/pipeline"
)

// Finalize unmounts the ISO, ejects the USB, and cleans up temp files.
type Finalize struct{}

func (Finalize) Key() string         { return "finalizing" }
func (Finalize) Name() string        { return "Finalizing" }
func (Finalize) HasProgress() bool   { return false }

func (Finalize) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 500*time.Millisecond, 2)
	}

	// Unmount ISO
	if state.ISOMountPath != "" {
		e.Emit(core.Event{Type: "log", Message: "Unmounting ISO..."})
		exec.CommandContext(ctx, "/usr/bin/hdiutil", "detach", state.ISOMountPath, "-force").Run()
		state.ISOMountPath = ""
	}

	// Clean up temp WIM files
	if state.SWMTempDir != "" {
		e.Emit(core.Event{Type: "log", Message: "Cleaning up temp files..."})
		os.RemoveAll(state.SWMTempDir)
		state.SWMTempDir = ""
	}

	// Eject USB
	e.Emit(core.Event{Type: "log", Message: "Ejecting USB..."})
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", state.TargetDisk)
	if out, err := cmd.CombinedOutput(); err != nil {
		outStr := strings.TrimSpace(string(out))
		e.Emit(core.Event{Type: "log", Message: "Warning: eject failed: " + outStr})
		// Don't fail for eject errors
	}

	e.Emit(core.Event{Type: "log", Message: "Windows USB created successfully!"})
	return nil
}
