package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/kyleaupton/flashit/internal/core"
	"github.com/kyleaupton/flashit/internal/iso"
	"github.com/kyleaupton/flashit/internal/pipeline"
)

// MountISO mounts the Windows ISO file.
type MountISO struct{}

func (MountISO) Key() string       { return "mounting-iso" }
func (MountISO) Name() string      { return "Mounting ISO" }
func (MountISO) HasProgress() bool { return false }

func (MountISO) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 1*time.Second, 3)
	}

	// Detach any existing mount of this ISO first
	if devEntry := iso.DetachExisting(ctx, state.ISOPath); devEntry != "" {
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("Detached stale mount at %s", devEntry)})
	}

	e.Emit(core.Event{Type: "log", Message: "Mounting Windows ISO..."})

	result, err := iso.Mount(ctx, state.ISOPath)
	if err != nil {
		return err
	}

	state.ISOMountResult = result
	state.ISOMountPath = result.MountPath
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("ISO mounted at %s", result.MountPath)})
	return nil
}

// Cleanup unmounts the ISO.
func (MountISO) Cleanup(ctx context.Context, state *FlashContext, e core.Executor) error {
	if state.ISOMountResult != nil {
		e.Emit(core.Event{Type: "log", Message: "Unmounting ISO..."})
		iso.Unmount(ctx, state.ISOMountResult)
		state.ISOMountResult = nil
		state.ISOMountPath = ""
	}
	return nil
}
