package fs

import "context"

type ProgressFunc func(wrote int64, total int64)

// FileOps abstracts file operations that may need platform-specific handling.
type FileOps interface {
	RawWriteToBlockDevice(ctx context.Context, srcPath string, devicePath string, onProgress ProgressFunc) error
}

var defaultFS FileOps = platformFS()

func SetFileOps(f FileOps) {
	if f != nil {
		defaultFS = f
	}
}

func RawWriteToBlockDevice(ctx context.Context, srcPath string, devicePath string, onProgress ProgressFunc) error {
	return defaultFS.RawWriteToBlockDevice(ctx, srcPath, devicePath, onProgress)
}

