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

func (l linuxDiskOps) WriteISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	return errors.New("not implemented on linux yet")
}
