//go:build darwin && !cgo

package macos

import (
	"context"
	"errors"
)

type xpcClient struct{}

func NewClient() Client { return nil }

func (c *xpcClient) EnsureReady(ctx context.Context) error { return nil }
func (c *xpcClient) WriteLinuxISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	return errors.New("WriteLinuxISO requires CGO build")
}
