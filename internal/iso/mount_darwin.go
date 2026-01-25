//go:build darwin

package iso

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"howett.net/plist"
)

type darwinMounter struct{}

func platformMounter() Mounter { return &darwinMounter{} }

func (m *darwinMounter) IsSupported() bool { return true }

// hdiutil info plist structures
type hdiutilInfo struct {
	Images []hdiutilImage `plist:"images"`
}

type hdiutilImage struct {
	ImagePath      string                `plist:"image-path"`
	SystemEntities []hdiutilSystemEntity `plist:"system-entities"`
}

type hdiutilSystemEntity struct {
	DevEntry   string `plist:"dev-entry"`
	MountPoint string `plist:"mount-point"`
}

func (m *darwinMounter) Mount(ctx context.Context, isoPath string) (*MountResult, error) {
	cmd := exec.CommandContext(ctx, "/usr/bin/hdiutil", "attach", "-readonly", "-nobrowse", "-mountrandom", "/tmp", isoPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("hdiutil attach failed: %s: %w", string(out), err)
	}

	mountPoint, devEntry, err := parseHdiutilOutput(string(out))
	if err != nil {
		return nil, err
	}

	return &MountResult{
		MountPath: mountPoint,
		DevEntry:  devEntry,
	}, nil
}

func (m *darwinMounter) Unmount(ctx context.Context, result *MountResult) error {
	if result == nil {
		return nil
	}

	// Prefer DevEntry for unmounting, fall back to MountPath
	target := result.DevEntry
	if target == "" {
		target = result.MountPath
	}
	if target == "" {
		return nil
	}

	detachCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(detachCtx, "/usr/bin/hdiutil", "detach", target, "-force")
	cmd.Run() // Ignore errors - unmount is best effort
	return nil
}

func (m *darwinMounter) DetachExisting(ctx context.Context, isoPath string) string {
	cmd := exec.CommandContext(ctx, "/usr/bin/hdiutil", "info", "-plist")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	var info hdiutilInfo
	if _, err := plist.Unmarshal(out, &info); err != nil {
		return ""
	}

	for _, img := range info.Images {
		if img.ImagePath == isoPath {
			for _, entity := range img.SystemEntities {
				if entity.DevEntry != "" {
					detachCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					exec.CommandContext(detachCtx, "/usr/bin/hdiutil", "detach", entity.DevEntry, "-force").Run()
					return entity.DevEntry
				}
			}
		}
	}
	return ""
}

// parseHdiutilOutput extracts the mount point and device entry from hdiutil attach output.
func parseHdiutilOutput(out string) (mountPoint string, devEntry string, err error) {
	// Output format varies, but typically:
	// /dev/disk13         	                               	/private/tmp/dmg.kQ9T6B
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			// First field is usually the device
			if strings.HasPrefix(fields[0], "/dev/disk") {
				devEntry = fields[0]
			}
		}
		if len(fields) >= 2 {
			lastField := fields[len(fields)-1]
			// Mount point will be under /tmp, /private/tmp, or /Volumes
			if strings.HasPrefix(lastField, "/tmp") ||
				strings.HasPrefix(lastField, "/private/tmp") ||
				strings.HasPrefix(lastField, "/Volumes") {
				mountPoint = lastField
			}
		}
	}

	if mountPoint == "" {
		return "", "", fmt.Errorf("could not parse mount point from hdiutil output: %s", out)
	}

	return mountPoint, devEntry, nil
}
