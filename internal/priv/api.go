package priv

import "context"

// ProgressFunc reports progress for long-running operations.
// total may be zero if unknown.
type ProgressFunc func(wrote, total int64)

// DiskOps defines privileged disk operations.
type DiskOps interface {
	UnmountDisk(ctx context.Context, device string) (string, error)
	RawWrite(ctx context.Context, isoPath string, rawDevice string, progress ProgressFunc) error
	EjectDisk(ctx context.Context, device string) (string, error)
}

// PrivilegedService provides access to privileged operations.
// EnsureReady should perform any elevation/installation needed to execute
// privileged operations for this session.
type PrivilegedService interface {
	EnsureReady(ctx context.Context) error
	Disk() DiskOps
	Shutdown(ctx context.Context) error
}

var defaultService PrivilegedService = platformService()

// NewService returns a process-wide singleton service implementation so that
// elevation/authorization can be requested once and reused by all steps.
func NewService() PrivilegedService { return defaultService }
