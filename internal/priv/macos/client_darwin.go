//go:build darwin && cgo && osinstallxpc

package macos

import (
	"boot-builder/internal/priv/macos/xpc"
	"context"
	"errors"
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

func (c *xpcClientWrapper) WriteLinuxISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	// Convert macos.ProgressFunc to xpc.ProgressFunc
	var xpcProgress xpc.ProgressFunc
	if progress != nil {
		xpcProgress = func(written, total uint64) {
			progress(written, total)
		}
	}
	err := c.client.WriteLinuxISO(ctx, isoPath, device, xpcProgress)
	// Convert xpc.ErrCancelled to macos.ErrCancelled for consistent error checking
	if errors.Is(err, xpc.ErrCancelled) {
		return ErrCancelled
	}
	return err
}

func (c *xpcClientWrapper) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	return c.client.FormatDisk(ctx, device, filesystem, volumeName)
}

func (c *xpcClientWrapper) CancelCurrentOperation() {
	c.client.CancelCurrentOperation()
}
