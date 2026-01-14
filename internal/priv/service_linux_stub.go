//go:build linux

package priv

import (
	"context"
	"errors"
)

type linuxService struct{}

func platformService() PrivilegedService { return &linuxService{} }

func (s *linuxService) EnsureReady(ctx context.Context) error {
	return errors.New("privileged service not implemented on linux yet")
}
func (s *linuxService) Disk() DiskOps                      { return linuxDiskOps{} }
func (s *linuxService) Shutdown(ctx context.Context) error { return nil }

type linuxDiskOps struct{}

func (l linuxDiskOps) UnmountDisk(ctx context.Context, device string) (string, error) {
	return "", errors.New("not implemented on linux yet")
}
func (l linuxDiskOps) RawWrite(ctx context.Context, isoPath string, rawDevice string, progress ProgressFunc) error {
	return errors.New("not implemented on linux yet")
}
func (l linuxDiskOps) EjectDisk(ctx context.Context, device string) (string, error) {
	return "", errors.New("not implemented on linux yet")
}
