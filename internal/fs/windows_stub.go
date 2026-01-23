//go:build windows

package fs

import (
	"context"
	"errors"
)

type windowsFS struct{}

func platformFS() FileOps { return &windowsFS{} }

func (w *windowsFS) CopyDir(ctx context.Context, src, dst string, opts CopyOptions, onProgress ProgressFunc) error {
	return errors.New("fs.CopyDir not implemented on windows yet")
}
