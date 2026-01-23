//go:build windows

package priv

import (
	"context"
	"errors"
)

type windowsService struct{}

func platformService() PrivilegedService { return &windowsService{} }

func (s *windowsService) EnsureReady(ctx context.Context) error {
	return errors.New("privileged service not implemented on windows yet")
}
func (s *windowsService) Disk() DiskOps                      { return windowsDiskOps{} }
func (s *windowsService) Shutdown(ctx context.Context) error { return nil }

type windowsDiskOps struct{}

func (w windowsDiskOps) WriteISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	return errors.New("not implemented on windows yet")
}

func (w windowsDiskOps) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	return errors.New("not implemented on windows yet")
}
