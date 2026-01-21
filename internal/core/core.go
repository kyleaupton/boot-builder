package core

import (
	"context"
	"time"
)

// DryRun controls whether real disk operations are performed.
// When true, installers should return mock steps and drives are simulated.
// Set via DRY_RUN environment variable.
var DryRun bool

type OSFamily string

const (
	OSWindows OSFamily = "windows"
	OSLinux   OSFamily = "linux"
	OSMacOS   OSFamily = "macos"
)

type HostInfo struct {
	OS       OSFamily
	Arch     string
	Version  string
	HasAdmin bool
}

type Capability struct {
	Supported bool
	Reasons   []string
}

type Target struct {
	Family  OSFamily
	Version string
	Arch    string
}

type SourceSpec struct {
	URL      string
	Checksum string
	Local    string
}

type SourceMode string

const (
	SourceModeSupply   SourceMode = "supply"
	SourceModeDownload SourceMode = "download"
	SourceModeBoth     SourceMode = "both"
)

type CreateRequest struct {
	Target  Target
	DriveID string
	Options map[string]any
	Source  SourceSpec
}

type Plan struct {
	ID    string
	Name  string
	Steps []Step
	Meta  map[string]any
}

type Step interface {
	Name() string
	Run(ctx context.Context, e Executor) error
	Estimate() time.Duration
}

type Installer interface {
	ID() string
	Name() string
	Targets() []Target
	AllowedSources() SourceMode
	ValidateHost(ctx context.Context, host HostInfo) Capability
	Plan(ctx context.Context, req CreateRequest) (*Plan, error)
}

type Event struct {
	JobID   string
	Type    string
	Message string
	Step    string
	Percent float64
	Error   string
}

type Executor interface {
	Emit(ev Event)
}
