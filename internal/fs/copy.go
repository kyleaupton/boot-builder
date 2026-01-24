package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CopyFileOptions configures single file copy behavior.
type CopyFileOptions struct {
	SyncAfter        bool          // Sync to disk after copy completes (default: true if zero value)
	SyncEvery        int64         // Sync after N bytes written (0 = no intermediate syncs)
	ProgressInterval time.Duration // Throttle progress callbacks (0 = every chunk)
	BufferSize       int           // Read/write buffer size (default 2MB)
}

// CopyDirOptions configures directory copy behavior.
type CopyDirOptions struct {
	// Filter returns true if a file should be copied, false to skip.
	// Receives the relative path (from src root) and file info.
	// If nil, all files are copied.
	Filter func(relPath string, info os.FileInfo) bool

	SyncAfter        bool          // Sync each file to disk after copy completes
	SyncEvery        int64         // Sync after N bytes written per file (0 = no intermediate syncs)
	ProgressInterval time.Duration // Throttle progress callbacks (0 = every chunk)
	BufferSize       int           // Read/write buffer size (default 2MB)
}

// CopyProgress reports copy progress.
type CopyProgress struct {
	Written int64 // Bytes written so far
	Total   int64 // Total bytes (-1 if unknown)
}

const defaultBufferSize = 2 * 1024 * 1024 // 2MB

// CopyFile copies a single file from src to dst with configurable sync and progress throttling.
// The onProgress callback receives progress updates and should return true to continue, false to cancel.
// If onProgress is nil, no progress reporting occurs.
func CopyFile(ctx context.Context, src, dst string, opts CopyFileOptions, onProgress func(CopyProgress) bool) error {
	// Apply defaults
	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = defaultBufferSize
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get file size for progress reporting
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := srcInfo.Size()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, bufSize)
	var written int64
	var lastSyncAt int64
	var lastProgressAt time.Time

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, readErr := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			written += int64(n)

			// Intermediate sync if configured
			if opts.SyncEvery > 0 && written-lastSyncAt >= opts.SyncEvery {
				if err := dstFile.Sync(); err != nil {
					return err
				}
				lastSyncAt = written
			}

			// Throttled progress callback
			if onProgress != nil {
				shouldEmit := opts.ProgressInterval == 0 ||
					time.Since(lastProgressAt) >= opts.ProgressInterval ||
					written == totalSize // Always emit on completion

				if shouldEmit {
					if !onProgress(CopyProgress{Written: written, Total: totalSize}) {
						return context.Canceled
					}
					lastProgressAt = time.Now()
				}
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	// Final sync
	if opts.SyncAfter {
		if err := dstFile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

// CopyDir recursively copies files from src to dst with progress reporting.
// The onProgress callback receives cumulative progress across all files and should return true to continue, false to cancel.
// If onProgress is nil, no progress reporting occurs.
func CopyDir(ctx context.Context, src, dst string, opts CopyDirOptions, onProgress func(CopyProgress) bool) error {
	// First pass: calculate total size of files to copy
	var totalBytes int64
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Apply filter
		if opts.Filter != nil && !opts.Filter(relPath, info) {
			return nil
		}

		totalBytes += info.Size()
		return nil
	})
	if err != nil {
		return err
	}

	// Second pass: copy files with progress
	var writtenBytes int64
	var lastProgressAt time.Time

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Apply filter
		if opts.Filter != nil && !opts.Filter(relPath, info) {
			return nil
		}

		// Copy the file with progress
		fileOpts := CopyFileOptions{
			SyncAfter:  opts.SyncAfter,
			SyncEvery:  opts.SyncEvery,
			BufferSize: opts.BufferSize,
			// Don't throttle individual file progress - we throttle at directory level
			ProgressInterval: 0,
		}

		fileSize := info.Size()
		err = CopyFile(ctx, path, dstPath, fileOpts, func(p CopyProgress) bool {
			if onProgress == nil {
				return true
			}

			// Throttle progress at directory level
			shouldEmit := opts.ProgressInterval == 0 ||
				time.Since(lastProgressAt) >= opts.ProgressInterval ||
				(writtenBytes+p.Written == totalBytes) // Always emit on completion

			if shouldEmit {
				if !onProgress(CopyProgress{Written: writtenBytes + p.Written, Total: totalBytes}) {
					return false
				}
				lastProgressAt = time.Now()
			}
			return true
		})
		if err != nil {
			return err
		}
		writtenBytes += fileSize

		return nil
	})
}
