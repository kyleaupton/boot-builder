//go:build darwin && !cgo

package macos

import "context"

type xpcClient struct{}

func NewClient() Client { return nil }

func (c *xpcClient) EnsureReady(ctx context.Context) error                          { return nil }
func (c *xpcClient) UnmountDisk(ctx context.Context, device string) (string, error) { return "", nil }
func (c *xpcClient) RawWrite(ctx context.Context, isoPath string, rawDevice string, onProgress func(wrote, total int64)) error {
	return nil
}
func (c *xpcClient) EjectDisk(ctx context.Context, device string) (string, error) { return "", nil }
