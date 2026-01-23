//go:build !darwin

package steps

import (
	"boot-builder/internal/core"
	"context"
	"errors"
	"time"
)

// DarwinWriteWindowsISO is a stub for non-darwin platforms.
type DarwinWriteWindowsISO struct {
	ISOPath    string
	Device     string
	VolumeName string
}

func (d DarwinWriteWindowsISO) Name() string            { return "Create Windows USB" }
func (d DarwinWriteWindowsISO) Estimate() time.Duration { return 10 * time.Minute }
func (d DarwinWriteWindowsISO) Run(ctx context.Context, e core.Executor) error {
	return errors.New("Windows USB creation is only supported on macOS")
}
