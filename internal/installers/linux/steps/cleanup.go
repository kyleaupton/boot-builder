//go:build darwin

package steps

import (
	"context"
	"os"

	"boot-builder/internal/core"
)

// Cleanup removes temporary files created during the pipeline.
// This step always succeeds - cleanup errors are logged but not fatal.
type Cleanup struct{}

func (Cleanup) Key() string         { return "cleanup" }
func (Cleanup) Name() string        { return "Cleaning up" }
func (Cleanup) HasProgress() bool   { return false }

func (Cleanup) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if state.TempISOPath != "" {
		e.Emit(core.Event{Type: "log", Message: "Removing temporary ISO..."})
		if err := os.Remove(state.TempISOPath); err != nil {
			e.Emit(core.Event{Type: "log", Message: "Warning: failed to remove temp ISO: " + err.Error()})
			// Don't fail the step for cleanup errors
		}
		state.TempISOPath = ""
	}
	return nil
}
