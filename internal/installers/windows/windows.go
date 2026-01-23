package windows

import (
	"boot-builder/internal/core"
	"boot-builder/internal/drives"
	"boot-builder/internal/steps"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// FAT32 max file size is 4GB - 1 byte
	fat32MaxFileSize = 4*1024*1024*1024 - 1
	// Volume name for the USB drive
	defaultVolumeName = "YOURNAME"
)

// Windows is an installer for Windows ISOs on macOS.
// Windows ISOs are not hybrid images, so they cannot be written directly to USB.
// Instead, we format the USB as FAT32 and copy the files, splitting large WIM
// files if necessary (FAT32 has a 4GB file size limit).
type Windows struct{}

func (w Windows) ID() string   { return "windows" }
func (w Windows) Name() string { return "Windows" }

func (w Windows) Targets() []core.Target {
	return []core.Target{{Family: core.OSWindows}}
}

func (w Windows) AllowedSources() core.SourceMode {
	return core.SourceModeSupply
}

func (w Windows) ValidateHost(ctx context.Context, host core.HostInfo) core.Capability {
	// Windows USB creation is currently only supported on macOS
	if runtime.GOOS != "darwin" {
		return core.Capability{
			Supported: false,
			Reasons:   []string{"Windows USB creation is currently only supported on macOS"},
		}
	}
	return core.Capability{Supported: true}
}

func (w Windows) Plan(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
	// Validate the ISO exists
	if req.Source.Local == "" {
		return nil, errors.New("ISO path is required")
	}

	info, err := os.Stat(req.Source.Local)
	if err != nil {
		return nil, fmt.Errorf("cannot access ISO: %w", err)
	}
	if info.IsDir() {
		return nil, errors.New("ISO path is a directory, not a file")
	}

	// Dry-run mode: return simulated steps for UI testing
	if core.DryRun {
		return &core.Plan{
			ID:   "plan-windows-dryrun",
			Name: "Windows USB (dry-run)",
			Steps: []core.Step{
				steps.NoOp{Label: "Mounting ISO...", Delay: 1 * time.Second, Ticks: 3},
				steps.NoOp{Label: "Formatting USB as FAT32...", Delay: 2 * time.Second, Ticks: 5},
				steps.NoOp{Label: "Copying boot files...", Delay: 5 * time.Second, Ticks: 15},
				steps.NoOp{Label: "Processing install.wim...", Delay: 10 * time.Second, Ticks: 30},
				steps.NoOp{Label: "Ejecting disk...", Delay: 500 * time.Millisecond, Ticks: 2},
			},
		}, nil
	}

	if runtime.GOOS == "darwin" {
		return w.planDarwin(ctx, req)
	}

	// Stub for other platforms
	return &core.Plan{
		ID:   "plan-windows-stub",
		Name: "Windows USB (stub)",
		Steps: []core.Step{
			steps.NoOp{Label: "Windows USB creation (stub)", Delay: 500 * time.Millisecond},
		},
	}, nil
}

func (w Windows) planDarwin(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
	// Validate drive ID
	if req.DriveID == "" {
		return nil, errors.New("DriveID is required")
	}

	// Guard against targeting the boot drive
	if strings.HasSuffix(req.DriveID, "disk0") || req.DriveID == "/dev/disk0" {
		return nil, errors.New("refusing to target /dev/disk0")
	}

	// Normalize device path
	if !strings.HasPrefix(req.DriveID, "/dev/") {
		req.DriveID = "/dev/" + req.DriveID
	}

	// Verify the drive is removable
	removable, err := drives.ListRemovable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list removable drives: %w", err)
	}

	found := false
	for _, d := range removable {
		if d.Device == req.DriveID {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("DriveID not recognized as removable USB drive")
	}

	// Check if install.wim needs splitting by mounting ISO first and checking size
	// For now, we'll do this check at runtime in the step itself
	// This allows the plan to be created quickly without mounting

	// Determine volume name from options or use default
	volumeName := defaultVolumeName
	if name, ok := req.Options["volumeName"].(string); ok && name != "" {
		volumeName = name
	}

	return &core.Plan{
		ID:   "plan-windows-darwin",
		Name: "Windows USB (darwin)",
		Steps: []core.Step{
			steps.DarwinWriteWindowsISO{
				ISOPath:    req.Source.Local,
				Device:     req.DriveID,
				VolumeName: volumeName,
			},
		},
	}, nil
}

// wimNeedsSplit checks if a WIM file exceeds the FAT32 file size limit.
// This is exported for use by steps if needed.
func WimNeedsSplit(wimPath string) (bool, int64, error) {
	info, err := os.Stat(wimPath)
	if err != nil {
		return false, 0, err
	}
	return info.Size() > fat32MaxFileSize, info.Size(), nil
}

// findInstallWim locates the install.wim file in a mounted Windows ISO.
// Returns the path to install.wim or an error if not found.
func FindInstallWim(mountPoint string) (string, error) {
	// Windows ISOs typically have install.wim in sources/
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
