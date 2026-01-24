package windows

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"boot-builder/internal/core"
	"boot-builder/internal/drives"
	winsteps "boot-builder/internal/installers/windows/steps"
	"boot-builder/internal/pipeline"
	"boot-builder/internal/priv"
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
	// Skip validation in dry-run mode (file may not exist)
	if !core.DryRun {
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
	}

	if runtime.GOOS == "darwin" {
		return w.planDarwin(ctx, req)
	}

	return nil, errors.New("Windows USB creation is only supported on macOS currently")
}

func (w Windows) planDarwin(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
	// Determine volume name from options or use default
	volumeName := defaultVolumeName
	if name, ok := req.Options["volumeName"].(string); ok && name != "" {
		volumeName = name
	}

	state := &winsteps.FlashContext{
		ISOPath:    req.Source.Local,
		TargetDisk: req.DriveID,
		VolumeName: volumeName,
	}

	// Skip validation and privileged service in dry-run mode
	if !core.DryRun {
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
			state.TargetDisk = req.DriveID
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

		// Initialize privileged service
		svc := priv.NewService()
		if err := svc.EnsureReady(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize privileged service: %w", err)
		}
		state.PrivService = svc
	}

	p := pipeline.New(
		winsteps.MountISO{},
		winsteps.FormatUSB{},
		winsteps.AnalyzeWim{},
		winsteps.CopyFiles{},
		winsteps.SplitWim{},
		winsteps.CopyWim{},
		winsteps.Finalize{},
	)

	runnable := pipeline.Bind(p, state)

	return &core.Plan{
		ID:        "plan-windows-darwin",
		Name:      "Windows USB (darwin)",
		Runnable:  runnable,
		StepInfos: runnable.StepInfos(),
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
