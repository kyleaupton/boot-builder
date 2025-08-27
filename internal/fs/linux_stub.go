//go:build linux

package fs

import (
	"context"
	"errors"
)

type linuxFS struct{}

func platformFS() FileOps { return &linuxFS{} }

func (l *linuxFS) RawWriteToBlockDevice(ctx context.Context, srcPath string, devicePath string, onProgress ProgressFunc) error {
	return errors.New("fs.RawWriteToBlockDevice not implemented on linux yet")
}

