package macos

import "context"

// Client defines the XPC client interface to the privileged helper.
// This interface is shared across darwin and non-darwin builds.
//
// Design: The helper orchestrates Apple's disk tools (diskutil, hdiutil, asr)
// rather than attempting raw disk I/O. This works with macOS security model.
type Client interface {
	EnsureReady(ctx context.Context) error
	UnmountDisk(ctx context.Context, device string) (string, error)
	EjectDisk(ctx context.Context, device string) (string, error)
	// WriteLinuxISO writes a Linux ISO to disk using Apple's imaging pipeline
	// (hdiutil convert → asr restore). This handles unmount and eject internally.
	WriteLinuxISO(ctx context.Context, isoPath string, device string) error
}
