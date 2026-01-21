package steps

import (
	"boot-builder/internal/core"
	"context"
	"time"
)

type NoOp struct {
	Label string
	Delay time.Duration
	Ticks int // number of progress updates during execution (default: 10)
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
	ticks := n.Ticks
	if ticks <= 0 {
		ticks = 10
	}

	tickDuration := n.Estimate() / time.Duration(ticks)

	for i := 1; i <= ticks; i++ {
		select {
		case <-time.After(tickDuration):
			percent := float64(i) / float64(ticks) * 100
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
				Message: n.Name(),
			})
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	e.Emit(core.Event{Type: "log", Message: "Completed: " + n.Name()})
	return nil
}
