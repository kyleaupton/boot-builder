//go:build darwin

package steps

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"

	"boot-builder/internal/core"
	"boot-builder/internal/fs"
	"boot-builder/internal/logger"
	"boot-builder/internal/priv"
	"boot-builder/internal/wim"
)

const (
	// FAT32 max file size is 4GB - 1 byte
	fat32MaxFileSize = 4*1024*1024*1024 - 1
)

// DarwinWriteWindowsISO creates a bootable Windows USB on macOS.
// This is a composite step that handles the entire pipeline:
// 1. Mount Windows ISO
// 2. Format USB as FAT32 (privileged)
// 3. Copy boot files
// 4. Split WIM if needed, then copy
// 5. Cleanup and eject
type DarwinWriteWindowsISO struct {
	ISOPath    string
	Device     string
	VolumeName string
}

func (d DarwinWriteWindowsISO) Name() string { return "Create Windows USB" }
func (d DarwinWriteWindowsISO) Estimate() time.Duration {
	return 10 * time.Minute // Varies widely based on ISO size and USB speed
}

func (d DarwinWriteWindowsISO) Run(ctx context.Context, e core.Executor) error {
	logger.Info("windows flash starting", "iso", d.ISOPath, "device", d.Device)

	if runtime.GOOS != "darwin" {
		return errors.New("this step only runs on macOS")
	}

	if d.ISOPath == "" || d.Device == "" {
		return errors.New("missing ISOPath or Device")
	}

	// Safety check
	if d.Device == "/dev/disk0" || strings.HasSuffix(d.Device, "disk0") {
		return errors.New("refusing to write to /dev/disk0")
	}

	volumeName := d.VolumeName
	if volumeName == "" {
		volumeName = "YOURNAME"
	}

	// Step 1: Mount the Windows ISO
	logger.Debug("mounting ISO", "path", d.ISOPath)
	e.Emit(core.Event{Type: "log", Message: "Mounting Windows ISO..."})
	mountPoint, err := mountISO(ctx, d.ISOPath)
	if err != nil {
		logger.Error("ISO mount failed", "error", err)
		return fmt.Errorf("failed to mount ISO: %w", err)
	}
	defer unmountISO(ctx, mountPoint)
	logger.Debug("ISO mounted", "mountPoint", mountPoint)
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("ISO mounted at %s", mountPoint)})

	// Step 2: Format USB as FAT32 (privileged)
	logger.Debug("initializing privileged service")
	e.Emit(core.Event{Type: "log", Message: "Formatting USB as FAT32..."})
	svc := priv.NewService()
	if err := svc.EnsureReady(ctx); err != nil {
		logger.Error("privileged service failed", "error", err)
		return fmt.Errorf("failed to initialize privileged service: %w", err)
	}

	logger.Debug("formatting disk", "device", d.Device, "filesystem", "FAT32", "volume", volumeName)
	if err := svc.Disk().FormatDisk(ctx, d.Device, "FAT32", volumeName); err != nil {
		logger.Error("disk format failed", "error", err)
		return fmt.Errorf("failed to format USB: %w", err)
	}
	logger.Debug("disk formatted successfully")
	e.Emit(core.Event{Type: "log", Message: "USB formatted successfully"})

	// Give the system time to mount the newly formatted volume
	time.Sleep(2 * time.Second)

	// Find the mount point of the formatted USB
	usbMountPoint, err := findUSBMountPoint(ctx, d.Device)
	if err != nil {
		logger.Error("failed to find USB mount point", "error", err)
		return fmt.Errorf("failed to find USB mount point: %w", err)
	}
	logger.Debug("USB mount point found", "mountPoint", usbMountPoint)
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("USB mounted at %s", usbMountPoint)})

	// Step 3: Find install.wim and check if it needs splitting
	installWimPath, err := findInstallWim(mountPoint)
	if err != nil {
		logger.Error("failed to find install.wim", "error", err)
		return fmt.Errorf("failed to find install.wim: %w", err)
	}
	logger.Debug("found install.wim", "path", installWimPath)

	wimInfo, err := os.Stat(installWimPath)
	if err != nil {
		logger.Error("failed to stat install.wim", "error", err)
		return fmt.Errorf("failed to stat install.wim: %w", err)
	}

	needsSplit := wimInfo.Size() > fat32MaxFileSize
	logger.Info("install.wim analysis", "size", wimInfo.Size(), "needsSplit", needsSplit)
	e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("install.wim size: %.2f GB (needs split: %v)",
		float64(wimInfo.Size())/1e9, needsSplit)})

	// Step 4: Copy files (skipping any > FAT32 limit, which will be handled via WIM split)
	logger.Info("copying files", "src", mountPoint, "dst", usbMountPoint)
	e.Emit(core.Event{Type: "log", Message: "Copying files..."})
	copyOpts := fs.CopyDirOptions{
		Filter: func(relPath string, info os.FileInfo) bool {
			if info.Size() > fat32MaxFileSize {
				e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("Skipping large file %s (%.2f GB)", relPath, float64(info.Size())/1e9)})
				return false
			}
			return true
		},
		SyncAfter:        true,
		ProgressInterval: 250 * time.Millisecond,
	}
	if err := fs.CopyDir(ctx, mountPoint, usbMountPoint, copyOpts, func(p fs.CopyProgress) bool {
		if p.Total > 0 {
			percent := float64(p.Written) * 100.0 / float64(p.Total)
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
				Message: fmt.Sprintf("Copying: %.1f / %.1f GB", float64(p.Written)/1e9, float64(p.Total)/1e9),
			})
		}
		return true
	}); err != nil {
		logger.Error("failed to copy files", "error", err)
		return fmt.Errorf("failed to copy files: %w", err)
	}
	logger.Info("files copied successfully")

	// Step 5: Split and copy install.wim if it exceeded FAT32 limit
	if needsSplit {
		logger.Info("splitting WIM", "src", installWimPath, "size", wimInfo.Size())
		e.Emit(core.Event{Type: "log", Message: "Splitting install.wim for FAT32 compatibility..."})
		if err := splitAndCopyWim(ctx, installWimPath, usbMountPoint, e); err != nil {
			logger.Error("WIM split failed", "error", err)
			return fmt.Errorf("failed to split/copy WIM: %w", err)
		}
		logger.Debug("WIM split and copy completed")
	}

	// Step 6: Eject USB
	logger.Debug("ejecting USB", "device", d.Device)
	e.Emit(core.Event{Type: "log", Message: "Ejecting USB..."})
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", d.Device)
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Warn("eject failed", "output", string(out), "error", err)
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("Warning: eject failed: %s", string(out))})
		// Don't fail the whole operation for eject failure
	}

	logger.Info("windows USB created successfully", "device", d.Device)
	e.Emit(core.Event{Type: "log", Message: "Windows USB created successfully!"})
	return nil
}

// mountISO mounts a Windows ISO and returns the mount point.
func mountISO(ctx context.Context, isoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "/usr/bin/hdiutil", "attach", "-readonly", "-nobrowse", "-mountrandom", "/tmp", isoPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("hdiutil attach failed: %s: %w", string(out), err)
	}

	// Parse output to find mount point
	// Output format varies, but mount point is the last field starting with /
	// Example: /dev/disk13         	                               	/private/tmp/dmg.kQ9T6B
	lines := strings.Split(string(out), "\n")
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

	return "", fmt.Errorf("could not parse mount point from hdiutil output: %s", string(out))
}

// unmountISO unmounts a mounted ISO.
func unmountISO(ctx context.Context, mountPoint string) {
	exec.CommandContext(ctx, "/usr/bin/hdiutil", "detach", mountPoint, "-force").Run()
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

// findInstallWim locates the install.wim file in a mounted Windows ISO.
func findInstallWim(mountPoint string) (string, error) {
	candidates := []string{
		filepath.Join(mountPoint, "sources", "install.wim"),
		filepath.Join(mountPoint, "Sources", "install.wim"),
		filepath.Join(mountPoint, "SOURCES", "install.wim"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("install.wim not found in ISO")
}

// splitAndCopyWim splits a WIM file and copies the parts to the USB.
func splitAndCopyWim(ctx context.Context, wimPath, usbRoot string, e core.Executor) error {
	// Create temp directory for split files
	tempDir := filepath.Join(os.TempDir(), "bootbuilder-wim-"+uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Split the WIM file
	splitPrefix := filepath.Join(tempDir, "install.swm")
	opts := wim.SplitOptions{
		PartSizeMiB: 3800, // 3800 MiB parts for safety margin under FAT32's 4GB limit
	}

	progressCb := func(p wim.Progress) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if p.TotalBytes > 0 {
			percent := float64(p.DoneBytes) * 100.0 / float64(p.TotalBytes)
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
				Message: fmt.Sprintf("Splitting: %.1f%% (part %d)", percent, p.Part),
			})
		}
		return true
	}

	if err := wim.SplitWithProgress(ctx, wimPath, splitPrefix, opts, progressCb); err != nil {
		return fmt.Errorf("failed to split WIM: %w", err)
	}

	e.Emit(core.Event{Type: "log", Message: "WIM split complete, copying to USB..."})

	// Copy the split files to USB
	copyCb := func(done, total int64) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		if total > 0 {
			percent := float64(done) * 100.0 / float64(total)
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
				Message: fmt.Sprintf("Copying: %.1f / %.1f GB", float64(done)/1e9, float64(total)/1e9),
			})
		}
		return true
	}

	if err := wim.CopySWMs(ctx, tempDir, usbRoot, copyCb); err != nil {
		return fmt.Errorf("failed to copy split WIM files: %w", err)
	}

	return nil
}
