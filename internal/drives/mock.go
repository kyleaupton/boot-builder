package drives

import (
	"context"
	"time"
)

// MockProvider returns fake drives for UI testing.
type MockProvider struct {
	Drives []Drive
}

func (m MockProvider) IsSupported() bool { return true }

func (m MockProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	return m.Drives, nil
}

func (m MockProvider) Unmount(ctx context.Context, device string) error {
	return nil
}

func (m MockProvider) Eject(ctx context.Context, device string) error {
	return nil
}

func (m MockProvider) FindMountPoint(ctx context.Context, device string) (string, error) {
	for _, d := range m.Drives {
		if d.Device == device && len(d.Mountpoints) > 0 {
			return d.Mountpoints[0], nil
		}
	}
	return "/Volumes/MOCK", nil
}

func (m MockProvider) WaitForMount(ctx context.Context, device string, timeout time.Duration) (string, error) {
	return m.FindMountPoint(ctx, device)
}

// DefaultMockDrives returns realistic test data for UI development.
func DefaultMockDrives() []Drive {
	return []Drive{
		{
			Device:      "/dev/disk99",
			BSDName:     "disk99",
			SizeBytes:   8 * 1024 * 1024 * 1024, // 8GB
			Model:       "Virtual Test Drive",
			Vendor:      "Mock USB",
			Protocol:    "USB",
			IsRemovable: true,
			IsEjectable: true,
			Mountpoints: []string{"/Volumes/MOCK_USB"},
		},
		{
			Device:      "/dev/disk98",
			BSDName:     "disk98",
			SizeBytes:   32 * 1024 * 1024 * 1024, // 32GB
			Model:       "Simulated Flash Drive",
			Vendor:      "Test Co",
			Protocol:    "USB",
			IsRemovable: true,
			IsEjectable: true,
			Mountpoints: []string{"/Volumes/FLASH32"},
		},
	}
}
