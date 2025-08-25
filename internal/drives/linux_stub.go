//go:build linux

package drives

import (
	"context"
	"errors"
)

type linuxProvider struct{}

func platformProvider() Provider { return &linuxProvider{} }

func (p *linuxProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	// TODO: implement via /sys/block/* + udev (ID_BUS=usb), and parse /proc/mounts
	return nil, errors.New("drives: linux implementation not yet provided")
}
