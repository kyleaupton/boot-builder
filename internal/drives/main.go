package drives

import "context"

// Drive is a normalized, cross-platform description of a physical drive.
// For macOS, BSDName is like "disk3"; on Linux, it might be "sdb".
type Drive struct {
	Device      string // e.g. "/dev/disk3" (darwin), "/dev/sdb" (linux)
	BSDName     string // short name: "disk3" (darwin), "sdb" (linux), may be empty on Windows
	SizeBytes   uint64
	Model       string
	Vendor      string
	Serial      string
	Protocol    string // "USB", "SATA", "NVMe", "Thunderbolt", etc.
	IsRemovable bool
	IsEjectable bool
	Mountpoints []string // mountpoints for any child volumes (if known)
}

// Provider lists removable USB drives for the current platform.
type Provider interface {
	ListRemovable(ctx context.Context) ([]Drive, error)
}

// Default, platform-specific provider (set by per-OS file).
var defaultProvider Provider = platformProvider()

// ListRemovable returns removable USB whole media for the current OS using the default provider.
func ListRemovable(ctx context.Context) ([]Drive, error) {
	return defaultProvider.ListRemovable(ctx)
}

// SetProvider allows tests or callers to swap the provider implementation (e.g., a mock).
func SetProvider(p Provider) {
	if p == nil {
		return
	}
	defaultProvider = p
}

