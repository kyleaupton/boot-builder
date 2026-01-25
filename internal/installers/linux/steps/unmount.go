package steps

import (
	"context"
	"time"

	"github.com/kyleaupton/flashit/internal/core"
	"github.com/kyleaupton/flashit/internal/drives"
	"github.com/kyleaupton/flashit/internal/pipeline"
)

// Unmount unmounts all volumes on the target disk.
type Unmount struct{}

func (Unmount) Key() string       { return "unmounting" }
func (Unmount) Name() string      { return "Unmounting disk" }
func (Unmount) HasProgress() bool { return false }

func (Unmount) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 500*time.Millisecond, 3)
	}

	e.Emit(core.Event{Type: "log", Message: "Unmounting disk..."})

	if err := drives.Unmount(ctx, state.TargetDisk); err != nil {
		return err
	}

	e.Emit(core.Event{Type: "log", Message: "Disk unmounted"})
	return nil
}
