package ubuntu

import (
	"boot-builder/internal/core"
	"boot-builder/internal/drives"
	"boot-builder/internal/steps"
	"context"
	"errors"
	"runtime"
	"strings"
	"time"
)

type Ubuntu struct{}

func (u Ubuntu) ID() string   { return "linux.ubuntu" }
func (u Ubuntu) Name() string { return "Ubuntu" }
func (u Ubuntu) Targets() []core.Target {
	return []core.Target{{Family: core.OSLinux, Version: "24.04", Arch: "x86_64"}}
}

func (u Ubuntu) AllowedSources() core.SourceMode { return core.SourceModeSupply }

func (u Ubuntu) ValidateHost(ctx context.Context, host core.HostInfo) core.Capability {
	return core.Capability{Supported: true}
}

func (u Ubuntu) Plan(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
	if runtime.GOOS == "darwin" {
		// Guard against unsafe drives and non-removable targets
		if req.DriveID == "" {
			return nil, errors.New("DriveID is required")
		}
		if strings.HasSuffix(req.DriveID, "disk0") || req.DriveID == "/dev/disk0" {
			return nil, errors.New("refusing to target /dev/disk0")
		}
		if !strings.HasPrefix(req.DriveID, "/dev/") {
			req.DriveID = "/dev/" + req.DriveID
		}
		if ds, _ := drives.ListRemovable(ctx); len(ds) > 0 {
			ok := false
			for _, d := range ds {
				if d.Device == req.DriveID {
					ok = true
					break
				}
			}
			if !ok {
				return nil, errors.New("DriveID not recognized as removable USB drive")
			}
		}
		// Use single step that orchestrates Apple's imaging tools:
		// diskutil unmount → hdiutil convert → asr restore → eject
		p := &core.Plan{
			ID:   "plan-ubuntu-darwin",
			Name: "Ubuntu USB (darwin)",
			Steps: []core.Step{
				steps.DarwinWriteLinuxISO{ISOPath: req.Source.Local, Device: req.DriveID},
			},
		}
		return p, nil
	}

	p := &core.Plan{
		ID:   "plan-ubuntu-stub",
		Name: "Ubuntu USB (stub)",
		Steps: []core.Step{
			steps.NoOp{Label: "Download ISO (stub)", Delay: 500 * time.Millisecond},
			steps.NoOp{Label: "Write Image (stub)", Delay: 500 * time.Millisecond},
		},
	}
	return p, nil
}
