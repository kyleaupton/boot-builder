//go:build windows

package iso

import (
	"context"
	"errors"
)

var errMountNotImplemented = errors.New("iso: mount not implemented on windows")

type windowsMounter struct{}

func platformMounter() Mounter { return &windowsMounter{} }

func (m *windowsMounter) IsSupported() bool { return false }

func (m *windowsMounter) Mount(ctx context.Context, isoPath string) (*MountResult, error) {
	return nil, errMountNotImplemented
}

func (m *windowsMounter) Unmount(ctx context.Context, result *MountResult) error {
	return errMountNotImplemented
}

func (m *windowsMounter) DetachExisting(ctx context.Context, isoPath string) string {
	return ""
}
