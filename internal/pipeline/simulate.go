package pipeline

import (
	"context"
	"time"

	"boot-builder/internal/core"
)

// Simulate simulates a step for dry-run mode with delay and progress events.
// This allows testing the UI without performing real disk operations.
func Simulate(ctx context.Context, e core.Executor, delay time.Duration, ticks int) error {
	if ticks <= 0 {
		ticks = 10
	}

	tickDuration := delay / time.Duration(ticks)

	for i := 1; i <= ticks; i++ {
		select {
		case <-time.After(tickDuration):
			percent := float64(i) / float64(ticks) * 100
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
			})
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
