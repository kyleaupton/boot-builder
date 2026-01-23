package fs

import (
	"context"
	"os"
)

// ProgressFunc reports copy progress with bytes written and total bytes.
type ProgressFunc func(wrote int64, total int64)

// CopyOptions configures directory copy behavior.
type CopyOptions struct {
	// Filter returns true if a file should be copied, false to skip.
	// Receives the relative path (from src root) and file info.
	// If nil, all files are copied.
	Filter func(relPath string, info os.FileInfo) bool
}

// FileOps abstracts file operations that may need platform-specific handling.
type FileOps interface {
	CopyDir(ctx context.Context, src, dst string, opts CopyOptions, onProgress ProgressFunc) error
}

var defaultFS FileOps = platformFS()

func SetFileOps(f FileOps) {
	if f != nil {
		defaultFS = f
	}
}

// CopyDir recursively copies files from src to dst with progress reporting.
// Progress reports cumulative bytes written across all files.
func CopyDir(ctx context.Context, src, dst string, opts CopyOptions, onProgress ProgressFunc) error {
	return defaultFS.CopyDir(ctx, src, dst, opts, onProgress)
}

