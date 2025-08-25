//go:build windows

package drives

import (
	"context"
	"errors"
)

type windowsProvider struct{}

func platformProvider() Provider { return &windowsProvider{} }

func (p *windowsProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	// TODO: implement via SetupAPI (GUID_DEVINTERFACE_DISK) + IOCTL_STORAGE_QUERY_PROPERTY
	// and GetVolumeInformation / QueryDosDevice to map mountpoints.
	return nil, errors.New("drives: windows implementation not yet provided")
}
