//go:build darwin

package steps

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"howett.net/plist"

	"boot-builder/internal/core"
	"boot-builder/internal/pipeline"
)

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

// detachExistingMount checks if the ISO is already mounted and detaches it.
func detachExistingMount(ctx context.Context, isoPath string) string {
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

// MountISO mounts the Windows ISO file.
type MountISO struct{}

func (MountISO) Key() string         { return "mounting-iso" }
func (MountISO) Name() string        { return "Mounting ISO" }
func (MountISO) HasProgress() bool   { return false }

func (MountISO) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 1*time.Second, 3)
	}

	// Detach any existing mount of this ISO first
	if devEntry := detachExistingMount(ctx, state.ISOPath); devEntry != "" {
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("Detached stale mount at %s", devEntry)})
	}

	e.Emit(core.Event{Type: "log", Message: "Mounting Windows ISO..."})

	cmd := exec.CommandContext(ctx, "/usr/bin/hdiutil", "attach", "-readonly", "-nobrowse", "-mountrandom", "/tmp", state.ISOPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hdiutil attach failed: %s: %w", string(out), err)
	}

	// Parse output to find mount point
	mountPoint, err := parseHdiutilOutput(string(out))
	if err != nil {
		return err
	}

	state.ISOMountPath = mountPoint
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("ISO mounted at %s", mountPoint)})
	return nil
}

// Cleanup unmounts the ISO.
func (MountISO) Cleanup(_ context.Context, state *FlashContext, e core.Executor) error {
	if state.ISOMountPath != "" {
		e.Emit(core.Event{Type: "log", Message: "Unmounting ISO..."})
		detachCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		exec.CommandContext(detachCtx, "/usr/bin/hdiutil", "detach", state.ISOMountPath, "-force").Run()
		state.ISOMountPath = ""
	}
	return nil
}

// parseHdiutilOutput extracts the mount point from hdiutil attach output.
func parseHdiutilOutput(out string) (string, error) {
	// Output format varies, but mount point is the last field starting with /
	// Example: /dev/disk13         	                               	/private/tmp/dmg.kQ9T6B
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			lastField := fields[len(fields)-1]
			// Mount point will be under /tmp, /private/tmp, or /Volumes
			if strings.HasPrefix(lastField, "/tmp") ||
				strings.HasPrefix(lastField, "/private/tmp") ||
				strings.HasPrefix(lastField, "/Volumes") {
				return lastField, nil
			}
		}
	}

	return "", fmt.Errorf("could not parse mount point from hdiutil output: %s", out)
}
