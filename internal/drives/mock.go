package drives

import "context"

// MockProvider returns fake drives for UI testing.
type MockProvider struct {
	Drives []Drive
}

func (m MockProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	return m.Drives, nil
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
