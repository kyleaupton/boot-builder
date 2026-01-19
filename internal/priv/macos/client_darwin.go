//go:build darwin && cgo && osinstallxpc

package macos

import (
	"boot-builder/internal/priv/macos/xpc"
	"context"
)

// xpcClientWrapper wraps the xpc.Client to implement the macos.Client interface.
type xpcClientWrapper struct {
	client *xpc.Client
}

// NewClient creates a new XPC client that communicates with the privileged helper.
func NewClient() Client {
	return &xpcClientWrapper{client: xpc.NewClient()}
}

func (c *xpcClientWrapper) EnsureReady(ctx context.Context) error {
	return c.client.EnsureReady(ctx)
}

func (c *xpcClientWrapper) UnmountDisk(ctx context.Context, device string) (string, error) {
	return c.client.UnmountDisk(ctx, device)
}

func (c *xpcClientWrapper) EjectDisk(ctx context.Context, device string) (string, error) {
	return c.client.EjectDisk(ctx, device)
}

func (c *xpcClientWrapper) WriteLinuxISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	// Convert macos.ProgressFunc to xpc.ProgressFunc
	var xpcProgress xpc.ProgressFunc
	if progress != nil {
		xpcProgress = func(written, total uint64) {
			progress(written, total)
		}
	}
	return c.client.WriteLinuxISO(ctx, isoPath, device, xpcProgress)
}
