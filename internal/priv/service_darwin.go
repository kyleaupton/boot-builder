//go:build darwin

package priv

import (
	"boot-builder/internal/fs"
	macosclient "boot-builder/internal/priv/macos"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type darwinService struct {
	client macosclient.Client
	once   sync.Once
	ready  error
}

func platformService() PrivilegedService { return &darwinService{} }

func (s *darwinService) EnsureReady(ctx context.Context) error {
	s.once.Do(func() {
		// Try to initialize XPC helper first.
		c := macosclient.NewClient()
		if c != nil {
			if err := c.EnsureReady(ctx); err == nil {
				s.client = c
				s.ready = nil
				return
			}
		}
		// Fallback: no helper yet; direct execution will be used.
		s.ready = nil
	})
	return s.ready
}

func (s *darwinService) Disk() DiskOps {
	if s.client != nil {
		return &darwinDiskOpsXPC{client: s.client}
	}
	return darwinDiskOps{}
}

func (s *darwinService) Shutdown(ctx context.Context) error { return nil }

type darwinDiskOps struct{}

func (d darwinDiskOps) UnmountDisk(ctx context.Context, device string) (string, error) {
	if strings.TrimSpace(device) == "" {
		return "", errors.New("device not specified")
	}
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "unmountDisk", "force", device)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (d darwinDiskOps) RawWrite(ctx context.Context, isoPath string, rawDevice string, progress ProgressFunc) error {
	if strings.TrimSpace(isoPath) == "" || strings.TrimSpace(rawDevice) == "" {
		return errors.New("missing ISOPath or Device")
	}
	if rawDevice == "/dev/disk0" || strings.HasSuffix(rawDevice, "disk0") {
		return errors.New("refusing to write to /dev/disk0")
	}
	// Prefer raw device node for speed if a non-raw device was provided
	dev := rawDevice
	if strings.HasPrefix(dev, "/dev/disk") {
		dev = strings.Replace(dev, "/dev/disk", "/dev/rdisk", 1)
	}
	// Normalize isoPath to absolute for logging/trace consistency
	if !filepath.IsAbs(isoPath) {
		if abs, err := filepath.Abs(isoPath); err == nil {
			isoPath = abs
		}
	}
	return fs.RawWriteToBlockDevice(ctx, isoPath, dev, func(wrote int64, total int64) {
		if progress != nil {
			progress(wrote, total)
		}
	})
}

func (d darwinDiskOps) EjectDisk(ctx context.Context, device string) (string, error) {
	if strings.TrimSpace(device) == "" {
		return "", errors.New("device not specified")
	}
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", device)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

type darwinDiskOpsXPC struct{ client macosclient.Client }

func (d *darwinDiskOpsXPC) UnmountDisk(ctx context.Context, device string) (string, error) {
	return d.client.UnmountDisk(ctx, device)
}

func (d *darwinDiskOpsXPC) RawWrite(ctx context.Context, isoPath string, rawDevice string, progress ProgressFunc) error {
	return d.client.RawWrite(ctx, isoPath, rawDevice, func(wrote, total int64) {
		if progress != nil {
			progress(wrote, total)
		}
	})
}

func (d *darwinDiskOpsXPC) EjectDisk(ctx context.Context, device string) (string, error) {
	return d.client.EjectDisk(ctx, device)
}
