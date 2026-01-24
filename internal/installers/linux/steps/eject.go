//go:build darwin

package steps

import (
	"context"
	"os/exec"
	"strings"

	"boot-builder/internal/core"
)

// Eject ejects the target disk after writing is complete.
// This does not require privileges for removable USB drives on macOS.
type Eject struct{}

func (Eject) Key() string         { return "ejecting" }
func (Eject) Name() string        { return "Ejecting disk" }
func (Eject) HasProgress() bool   { return false }

func (Eject) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	e.Emit(core.Event{Type: "log", Message: "Ejecting disk..."})

	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", state.TargetDisk)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))

	if err != nil {
		// Eject failure is not fatal - just log it
		e.Emit(core.Event{Type: "log", Message: "Warning: eject failed: " + outStr})
		return nil
	}

	e.Emit(core.Event{Type: "log", Message: outStr})
	return nil
}
