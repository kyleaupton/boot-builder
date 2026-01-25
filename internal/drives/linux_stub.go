//go:build linux

package drives

import (
	"context"
	"errors"
	"time"
)

var errNotImplemented = errors.New("drives: linux implementation not yet provided")

type linuxProvider struct{}

func platformProvider() Provider { return &linuxProvider{} }

func (p *linuxProvider) IsSupported() bool { return false }

func (p *linuxProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	return nil, errNotImplemented
}

func (p *linuxProvider) Unmount(ctx context.Context, device string) error {
	return errNotImplemented
}

func (p *linuxProvider) Eject(ctx context.Context, device string) error {
	return errNotImplemented
}

func (p *linuxProvider) FindMountPoint(ctx context.Context, device string) (string, error) {
	return "", errNotImplemented
}

func (p *linuxProvider) WaitForMount(ctx context.Context, device string, timeout time.Duration) (string, error) {
	return "", errNotImplemented
}

