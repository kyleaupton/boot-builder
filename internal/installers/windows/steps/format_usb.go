//go:build darwin

package steps

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"boot-builder/internal/core"
	"boot-builder/internal/pipeline"
)

// FormatUSB formats the target disk as FAT32.
type FormatUSB struct{}

func (FormatUSB) Key() string         { return "formatting" }
func (FormatUSB) Name() string        { return "Formatting USB as FAT32" }
func (FormatUSB) HasProgress() bool   { return false }

func (FormatUSB) Run(ctx context.Context, state *FlashContext, e core.Executor) error {
	if core.DryRun {
		return pipeline.Simulate(ctx, e, 2*time.Second, 5)
	}

	e.Emit(core.Event{Type: "log", Message: "Formatting USB as FAT32..."})

	if err := state.PrivService.Disk().FormatDisk(ctx, state.TargetDisk, "FAT32", state.VolumeName); err != nil {
		return fmt.Errorf("failed to format USB: %w", err)
	}

	e.Emit(core.Event{Type: "log", Message: "USB formatted successfully"})

	// Give the system time to mount the newly formatted volume
	time.Sleep(2 * time.Second)

	// Find the mount point of the formatted USB
	mountPoint, err := findUSBMountPoint(ctx, state.TargetDisk)
	if err != nil {
		return fmt.Errorf("failed to find USB mount point: %w", err)
	}

	state.USBMountPath = mountPoint
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("USB mounted at %s", mountPoint)})
	return nil
}

// findUSBMountPoint finds where the USB is mounted after formatting.
func findUSBMountPoint(ctx context.Context, device string) (string, error) {
	// diskutil info gives us the mount point
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "info", "-plist", device+"s1")
	out, err := cmd.Output()
	if err != nil {
		// Try without partition suffix
		cmd = exec.CommandContext(ctx, "/usr/sbin/diskutil", "info", "-plist", device)
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("diskutil info failed: %w", err)
		}
	}

	// Simple plist parsing for MountPoint
	outStr := string(out)
	if idx := strings.Index(outStr, "<key>MountPoint</key>"); idx != -1 {
		rest := outStr[idx:]
		if startIdx := strings.Index(rest, "<string>"); startIdx != -1 {
			rest = rest[startIdx+8:]
			if endIdx := strings.Index(rest, "</string>"); endIdx != -1 {
				return rest[:endIdx], nil
			}
		}
	}

	return "", errors.New("could not find USB mount point")
}
