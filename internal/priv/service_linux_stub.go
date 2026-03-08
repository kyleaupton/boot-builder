//go:build linux

package priv

import (
	"context"
	"sync"

	"github.com/kyleaupton/flashit/internal/logger"
	linuxclient "github.com/kyleaupton/flashit/internal/priv/linux"
)

type linuxService struct {
	client linuxclient.Client
	once   sync.Once
	ready  error
}

func platformService() PrivilegedService { return &linuxService{} }

func (s *linuxService) EnsureReady(ctx context.Context) error {
	s.once.Do(func() {
		logger.Debug("checking privileged helper status")

		c := linuxclient.NewClient()
		if c == nil {
			logger.Warn("Linux helper client not available")
			s.ready = nil
			return
		}

		if err := c.EnsureReady(ctx); err != nil {
			logger.Warn("Linux helper not ready", "error", err)
			s.ready = err
			return
		}

		logger.Debug("privileged helper ready via unix socket")
		s.client = c
		s.ready = nil
	})
	return s.ready
}

func (s *linuxService) Disk() DiskOps {
	if s.client != nil {
		return &linuxDiskOps{client: s.client}
	}
	return linuxDiskOpsFallback{}
}

func (s *linuxService) Shutdown(ctx context.Context) error {
	if s.client != nil {
		return s.client.Shutdown(ctx)
	}
	return nil
}

// linuxDiskOps uses the privileged helper for disk operations.
type linuxDiskOps struct {
	client linuxclient.Client
}

func (d *linuxDiskOps) WriteISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	var linuxProgress linuxclient.ProgressFunc
	if progress != nil {
		linuxProgress = func(written, total uint64) {
			progress(written, total)
		}
	}
	return d.client.WriteISO(ctx, isoPath, device, linuxProgress)
}

func (d *linuxDiskOps) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	return d.client.FormatDisk(ctx, device, filesystem, volumeName)
}

// linuxDiskOpsFallback is used when the helper is not available.
type linuxDiskOpsFallback struct{}

func (d linuxDiskOpsFallback) WriteISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	return linuxclient.ErrHelperNotRunning
}

func (d linuxDiskOpsFallback) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	return linuxclient.ErrHelperNotRunning
}
