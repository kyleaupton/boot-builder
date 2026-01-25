package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"boot-builder/internal/core"
	"boot-builder/internal/fs"
	"boot-builder/internal/pipeline"
)

// Prepare clones the ISO to a neutral location to bypass TCC restrictions.
// The privileged helper cannot read TCC-protected directories (like ~/Downloads)
// even when running as root. APFS clone is near-instant for large files.
type Prepare struct{}

func (Prepare) Key() string         { return "preparing" }
func (Prepare) Name() string        { return "Preparing ISO" }
func (Prepare) HasProgress() bool   { return false }

func (Prepare) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 1*time.Second, 5)
	}

	e.Emit(core.Event{Type: "log", Message: "Preparing ISO..."})

	tempPath := filepath.Join(fs.TempISODir, uuid.New().String()+".iso")
	if err := fs.CloneFile(state.ISOPath, tempPath); err != nil {
		return fmt.Errorf("failed to prepare ISO: %w", err)
	}

	state.TempISOPath = tempPath
	e.Emit(core.Event{Type: "log", Message: "ISO prepared"})
	return nil
}

// Cleanup removes the cloned ISO file.
func (Prepare) Cleanup(_ context.Context, state *FlashContext, e core.Executor) error {
	if state.TempISOPath != "" {
		e.Emit(core.Event{Type: "log", Message: "Cleaning up temporary ISO..."})
		if err := os.Remove(state.TempISOPath); err != nil && !os.IsNotExist(err) {
			e.Emit(core.Event{Type: "log", Message: "Warning: failed to remove temp ISO: " + err.Error()})
		}
		state.TempISOPath = ""
	}
	return nil
}
