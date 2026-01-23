//go:build linux

package fs

import (
	"context"
	"errors"
)

type linuxFS struct{}

func platformFS() FileOps { return &linuxFS{} }

func (l *linuxFS) CopyDir(ctx context.Context, src, dst string, opts CopyOptions, onProgress ProgressFunc) error {
	return errors.New("fs.CopyDir not implemented on linux yet")
}
