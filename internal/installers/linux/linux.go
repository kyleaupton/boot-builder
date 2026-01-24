package linux

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"boot-builder/internal/core"
	"boot-builder/internal/drives"
	linuxsteps "boot-builder/internal/installers/linux/steps"
	"boot-builder/internal/iso"
	"boot-builder/internal/pipeline"
	"boot-builder/internal/priv"
	"boot-builder/internal/steps"
)

// Linux is a generic installer for any hybrid ISO Linux distribution.
// It validates that the provided ISO is a hybrid image (can be written
// directly to USB) and creates a plan to write it using platform-native tools.
type Linux struct{}

func (l Linux) ID() string   { return "linux" }
func (l Linux) Name() string { return "Linux" }

func (l Linux) Targets() []core.Target {
	return []core.Target{{Family: core.OSLinux}}
}

func (l Linux) AllowedSources() core.SourceMode {
	return core.SourceModeSupply
}

func (l Linux) ValidateHost(ctx context.Context, host core.HostInfo) core.Capability {
	return core.Capability{Supported: true}
}

func (l Linux) Plan(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
	// Validate that the ISO is a hybrid image
	if err := iso.ValidateHybrid(req.Source.Local); err != nil {
		if errors.Is(err, iso.ErrNotISO9660) {
			return nil, fmt.Errorf("invalid ISO file: %w", err)
		}
		if errors.Is(err, iso.ErrNotHybrid) {
			return nil, fmt.Errorf("this ISO cannot be written directly to USB (non-hybrid image): %w", err)
		}
		if errors.Is(err, iso.ErrFileTooSmall) {
			return nil, fmt.Errorf("file too small to be a valid ISO: %w", err)
		}
		return nil, fmt.Errorf("failed to validate ISO: %w", err)
	}

	// Dry-run mode: return simulated steps for UI testing
	if core.DryRun {
		return &core.Plan{
			ID:   "plan-linux-dryrun",
			Name: "Linux USB (dry-run)",
			Steps: []core.Step{
				steps.NoOp{Label: "Unmounting disk...", Delay: 1 * time.Second, Ticks: 5},
				steps.NoOp{Label: "Writing ISO to USB...", Delay: 8 * time.Second, Ticks: 20},
				steps.NoOp{Label: "Verifying write...", Delay: 2 * time.Second, Ticks: 10},
				steps.NoOp{Label: "Ejecting disk...", Delay: 500 * time.Millisecond, Ticks: 2},
			},
			StepInfos: []core.StepInfo{
				{Key: "unmounting", Name: "Unmounting disk", HasProgress: false},
				{Key: "writing-iso", Name: "Writing ISO to USB", HasProgress: true},
				{Key: "verifying", Name: "Verifying write", HasProgress: true},
				{Key: "ejecting", Name: "Ejecting disk", HasProgress: false},
			},
		}, nil
	}

	if runtime.GOOS == "darwin" {
		return l.planDarwin(ctx, req)
	}

	// Stub for Linux/Windows
	return &core.Plan{
		ID:   "plan-linux-stub",
		Name: "Linux USB (stub)",
		Steps: []core.Step{
			steps.NoOp{Label: "Write Image (stub)", Delay: 500 * time.Millisecond},
		},
		StepInfos: []core.StepInfo{
			{Key: "writing", Name: "Writing image", HasProgress: false},
		},
	}, nil
}

func (l Linux) planDarwin(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
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

	// Initialize privileged service
	svc := priv.NewService()
	if err := svc.EnsureReady(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize privileged service: %w", err)
	}

	// Build the pipeline with typed context
	state := &linuxsteps.FlashContext{
		ISOPath:     req.Source.Local,
		TargetDisk:  req.DriveID,
		PrivService: svc,
	}

	p := pipeline.New[linuxsteps.FlashContext](
		linuxsteps.Prepare{},
		linuxsteps.Unmount{},
		linuxsteps.Write{},
		linuxsteps.Eject{},
		linuxsteps.Cleanup{},
	)

	runnable := pipeline.Bind(p, state)

	return &core.Plan{
		ID:        "plan-linux-darwin",
		Name:      "Linux USB (darwin)",
		Runnable:  runnable,
		StepInfos: runnable.StepInfos(),
	}, nil
}
