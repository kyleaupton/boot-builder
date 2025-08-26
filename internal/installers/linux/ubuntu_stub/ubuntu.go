package ubuntu_stub

import (
	"boot-builder/internal/core"
	"boot-builder/internal/steps"
	"context"
	"time"
)

type UbuntuStub struct{}

func (u UbuntuStub) ID() string   { return "linux.ubuntu.stub" }
func (u UbuntuStub) Name() string { return "Ubuntu (stub)" }
func (u UbuntuStub) Targets() []core.Target {
	return []core.Target{{Family: core.OSLinux, Version: "24.04", Arch: "x86_64"}}
}

func (u UbuntuStub) ValidateHost(ctx context.Context, host core.HostInfo) core.Capability {
	return core.Capability{Supported: true}
}

func (u UbuntuStub) Plan(ctx context.Context, req core.CreateRequest) (*core.Plan, error) {
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
