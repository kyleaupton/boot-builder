package macos

import (
	"context"
	"errors"
)

// ErrCancelled is returned when an operation is cancelled by the user.
var ErrCancelled = errors.New("operation cancelled")

// ProgressFunc is the callback type for write progress updates.
// bytesWritten: total bytes written so far
// totalBytes: total size of the ISO
type ProgressFunc func(bytesWritten, totalBytes uint64)

// Client defines the XPC client interface to the privileged helper.
// This interface is shared across darwin and non-darwin builds.
//
// Design: The helper orchestrates Apple's disk tools (diskutil, hdiutil, asr)
// rather than attempting raw disk I/O. This works with macOS security model.
type Client interface {
	EnsureReady(ctx context.Context) error
	// WriteLinuxISO writes a Linux ISO to disk using Disk Arbitration and direct I/O.
	// The progress callback receives (bytesWritten, totalBytes) updates during the write.
	// Pass nil if progress updates are not needed.
	WriteLinuxISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error
	// FormatDisk formats a disk with the specified filesystem and volume name.
	// Supported filesystems: FAT32, ExFAT, APFS, HFS+
	FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error
	// CancelCurrentOperation cancels the currently running operation (if any).
	// This is safe to call even if no operation is running.
	CancelCurrentOperation()
}
