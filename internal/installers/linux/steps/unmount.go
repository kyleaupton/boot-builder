//go:build darwin

package steps

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"boot-builder/internal/core"
	"boot-builder/internal/pipeline"
)

// Unmount unmounts all volumes on the target disk.
// This does not require privileges for removable USB drives on macOS.
type Unmount struct{}

func (Unmount) Key() string         { return "unmounting" }
func (Unmount) Name() string        { return "Unmounting disk" }
func (Unmount) HasProgress() bool   { return false }

func (Unmount) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 500*time.Millisecond, 3)
	}

	e.Emit(core.Event{Type: "log", Message: "Unmounting disk..."})

	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "unmountDisk", "force", state.TargetDisk)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	if err != nil {
		// Check if it's already unmounted (not an error)
		outLower := strings.ToLower(outStr)
		if strings.Contains(outLower, "not mounted") || strings.Contains(outLower, "already unmounted") {
			e.Emit(core.Event{Type: "log", Message: "Disk already unmounted"})
			return nil
		}
		return fmt.Errorf("failed to unmount %s: %w (%s)", state.TargetDisk, err, outStr)
	}

	e.Emit(core.Event{Type: "log", Message: outStr})
	return nil
}
