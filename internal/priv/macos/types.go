package macos

import "context"

// Client defines the XPC client interface to the privileged helper.
// This interface is shared across darwin and non-darwin builds.
type Client interface {
	EnsureReady(ctx context.Context) error
	UnmountDisk(ctx context.Context, device string) (string, error)
	RawWrite(ctx context.Context, isoPath string, rawDevice string, onProgress func(wrote, total int64)) error
	EjectDisk(ctx context.Context, device string) (string, error)
}
