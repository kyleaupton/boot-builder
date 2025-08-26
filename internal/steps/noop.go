package steps

import (
	"boot-builder/internal/core"
	"context"
	"time"
)

type NoOp struct {
	Label string
	Delay time.Duration
}

func (n NoOp) Name() string {
	if n.Label != "" {
		return n.Label
	}
	return "NoOp"
}
func (n NoOp) Estimate() time.Duration {
	if n.Delay > 0 {
		return n.Delay
	}
	return 1 * time.Second
}
func (n NoOp) Run(ctx context.Context, e core.Executor) error {
	// tiny sleep to simulate work
	select {
	case <-time.After(n.Estimate()):
	case <-ctx.Done():
		return ctx.Err()
	}
	e.Emit(core.Event{Type: "log", Message: "Completed: " + n.Name()})
	return nil
}
