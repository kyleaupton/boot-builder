//go:build windows

package drives

import (
	"context"
	"errors"
	"time"
)

var errNotImplemented = errors.New("drives: windows implementation not yet provided")

type windowsProvider struct{}

func platformProvider() Provider { return &windowsProvider{} }

func (p *windowsProvider) IsSupported() bool { return false }

func (p *windowsProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	return nil, errNotImplemented
}

func (p *windowsProvider) Unmount(ctx context.Context, device string) error {
	return errNotImplemented
}

func (p *windowsProvider) Eject(ctx context.Context, device string) error {
	return errNotImplemented
}

func (p *windowsProvider) FindMountPoint(ctx context.Context, device string) (string, error) {
	return "", errNotImplemented
}

func (p *windowsProvider) WaitForMount(ctx context.Context, device string, timeout time.Duration) (string, error) {
	return "", errNotImplemented
}

