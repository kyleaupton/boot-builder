//go:build windows

package fs

import (
	"context"
	"errors"
)

type windowsFS struct{}

func platformFS() FileOps { return &windowsFS{} }

func (w *windowsFS) RawWriteToBlockDevice(ctx context.Context, srcPath string, devicePath string, onProgress ProgressFunc) error {
	return errors.New("fs.RawWriteToBlockDevice not implemented on windows yet")
}

