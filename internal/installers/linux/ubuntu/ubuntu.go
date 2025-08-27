package ubuntu

import (
	"boot-builder/internal/core"
	"boot-builder/internal/steps"
	"context"
	"runtime"
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
		p := &core.Plan{
			ID:   "plan-ubuntu-darwin",
			Name: "Ubuntu USB (darwin)",
			Steps: []core.Step{
				steps.DarwinUnmountDisk{Device: req.DriveID},
				steps.DarwinRawWriteISO{ISOPath: req.Source.Local, Device: req.DriveID},
				steps.DarwinEjectDisk{Device: req.DriveID},
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
