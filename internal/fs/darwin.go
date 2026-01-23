//go:build darwin

package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type darwinFS struct{}

func platformFS() FileOps { return &darwinFS{} }

func (d *darwinFS) CopyDir(ctx context.Context, src, dst string, opts CopyOptions, onProgress ProgressFunc) error {
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
		written, err := copyFileWithProgress(ctx, path, dstPath, func(fileWrote int64) {
			if onProgress != nil {
				onProgress(writtenBytes+fileWrote, totalBytes)
			}
		})
		if err != nil {
			return err
		}
		writtenBytes += written

		return nil
	})
}

// copyFileWithProgress copies a single file and calls onProgress with bytes written for this file.
// Returns the total bytes written.
func copyFileWithProgress(ctx context.Context, src, dst string, onProgress func(wrote int64)) (int64, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0, err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	buf := make([]byte, 2*1024*1024) // 2MB buffer
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		n, readErr := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return written, writeErr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(written)
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return written, readErr
		}
	}

	return written, nil
}
