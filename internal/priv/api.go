package priv

import "context"

// DiskOps defines privileged disk operations.
type DiskOps interface {
	UnmountDisk(ctx context.Context, device string) (string, error)
	EjectDisk(ctx context.Context, device string) (string, error)
	// WriteISO writes a Linux ISO to a disk device.
	// On macOS, this uses Disk Arbitration to claim exclusive access and direct I/O.
	// On Linux, this would use dd with pkexec.
	// This is a long-running operation that handles unmount and eject internally.
	WriteISO(ctx context.Context, isoPath string, device string) error
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
