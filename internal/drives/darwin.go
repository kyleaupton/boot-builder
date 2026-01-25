//go:build darwin

package drives

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"howett.net/plist"
)

type darwinProvider struct{}

func platformProvider() Provider { return &darwinProvider{} }

// ---------- plist models ----------

// `diskutil list -plist`
type duList struct {
	AllDisksAndPartitions []struct {
		Device     string `plist:"DeviceIdentifier"` // "disk4"
		Partitions []struct {
			Device     string `plist:"DeviceIdentifier"` // "disk4s3"
			MountPoint string `plist:"MountPoint,omitempty"`
		} `plist:"Partitions"`
	} `plist:"AllDisksAndPartitions"`

	// The authoritative set of top-level whole disks
	WholeDisks []string `plist:"WholeDisks"` // e.g. ["disk4"]
}

// `diskutil info -plist diskN`
type duInfo struct {
	DeviceNode       string `plist:"DeviceNode"`       // "/dev/disk4"
	DeviceIdentifier string `plist:"DeviceIdentifier"` // "disk4"
	WholeDisk        bool   `plist:"WholeDisk"`

	RemovableMedia bool   `plist:"RemovableMedia"`
	Ejectable      bool   `plist:"Ejectable"`
	BusProtocol    string `plist:"BusProtocol"`    // "USB" (preferred)
	DeviceProtocol string `plist:"DeviceProtocol"` // sometimes present instead

	TotalSize uint64 `plist:"TotalSize"` // bytes
	Size      uint64 `plist:"Size"`      // fallback on older versions

	// Best-effort identity fields (often empty depending on device)
	DeviceModel         string `plist:"DeviceModel"`
	DeviceVendor        string `plist:"DeviceVendor"`
	DeviceSerial        string `plist:"DeviceSerial"`
	IORegistryEntryName string `plist:"IORegistryEntryName"` // e.g. "PNY USB 3.0 FD Media"
	MediaName           string `plist:"MediaName"`           // e.g. "USB 3.0 FD"
	MountPoint          string `plist:"MountPoint"`          // volume mount point (if mounted)
}

// ---------- public API ----------

func (p *darwinProvider) IsSupported() bool { return true }

func (p *darwinProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	// Always bound the external commands with a timeout if none provided.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}

	// 1) Enumerate topology & mount points
	lst, err := p.diskutilList(ctx)
	if err != nil {
		return nil, err
	}

	mounts := indexMountpoints(lst)

	// 2) For each whole disk, fetch detailed flags via `diskutil info -plist <diskN>`
	var out []Drive
	for _, bsd := range lst.WholeDisks {
		inf, err := p.diskutilInfo(ctx, bsd)
		if err != nil {
			// Non-fatal: skip disks that fail to query
			continue
		}
		if !inf.WholeDisk {
			continue
		}

		proto := firstNonEmpty(inf.BusProtocol, inf.DeviceProtocol)
		if !strings.EqualFold(proto, "USB") {
			continue
		}
		if !(inf.RemovableMedia || inf.Ejectable) {
			continue
		}

		size := inf.TotalSize
		if size == 0 {
			size = inf.Size
		}

		drv := Drive{
			Device:      preferDevPath(inf.DeviceNode, inf.DeviceIdentifier),
			BSDName:     inf.DeviceIdentifier,
			SizeBytes:   size,
			Model:       firstNonEmpty(inf.DeviceModel, inf.IORegistryEntryName, inf.MediaName),
			Vendor:      inf.DeviceVendor,
			Serial:      inf.DeviceSerial,
			Protocol:    proto,
			IsRemovable: inf.RemovableMedia,
			IsEjectable: inf.Ejectable,
			Mountpoints: mounts[bsd],
		}
		out = append(out, drv)
	}

	return out, nil
}

func (p *darwinProvider) Unmount(ctx context.Context, device string) error {
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "unmountDisk", "force", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		outLower := strings.ToLower(outStr)
		// Already unmounted is not an error
		if strings.Contains(outLower, "not mounted") || strings.Contains(outLower, "already unmounted") {
			return nil
		}
		return fmt.Errorf("failed to unmount %s: %w (%s)", device, err, outStr)
	}
	return nil
}

func (p *darwinProvider) Eject(ctx context.Context, device string) error {
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to eject %s: %w (%s)", device, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (p *darwinProvider) FindMountPoint(ctx context.Context, device string) (string, error) {
	// Try with s1 suffix first (first partition)
	candidates := []string{device + "s1", device}

	for _, dev := range candidates {
		inf, err := p.diskutilInfo(ctx, dev)
		if err != nil {
			continue
		}
		if inf.MountPoint != "" {
			return inf.MountPoint, nil
		}
	}

	return "", fmt.Errorf("no mount point found for %s", device)
}

func (p *darwinProvider) WaitForMount(ctx context.Context, device string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if mp, err := p.FindMountPoint(ctx, device); err == nil && mp != "" {
			return mp, nil
		}

		time.Sleep(pollInterval)
	}

	return "", fmt.Errorf("timeout waiting for %s to mount", device)
}

// ---------- helpers ----------

func (p *darwinProvider) diskutilList(ctx context.Context) (*duList, error) {
	raw, err := exec.CommandContext(ctx, "/usr/sbin/diskutil", "list", "-plist").Output()
	if err != nil {
		return nil, wrapExecErr("diskutil list -plist", err)
	}
	var lst duList
	if _, err := plist.Unmarshal(raw, &lst); err != nil {
		return nil, fmt.Errorf("parse diskutil list -plist: %w", err)
	}
	return &lst, nil
}

func (p *darwinProvider) diskutilInfo(ctx context.Context, diskID string) (*duInfo, error) {
	// diskutil accepts either "disk4" or "/dev/disk4"
	arg := diskID
	if !strings.HasPrefix(arg, "/dev/") {
		arg = "/dev/" + arg
	}
	raw, err := exec.CommandContext(ctx, "/usr/sbin/diskutil", "info", "-plist", arg).Output()
	if err != nil {
		return nil, wrapExecErr("diskutil info -plist "+diskID, err)
	}
	var inf duInfo
	if _, err := plist.Unmarshal(raw, &inf); err != nil {
		return nil, fmt.Errorf("parse diskutil info -plist %s: %w", diskID, err)
	}
	return &inf, nil
}

func indexMountpoints(lst *duList) map[string][]string {
	mp := make(map[string][]string, len(lst.AllDisksAndPartitions))
	for _, d := range lst.AllDisksAndPartitions {
		for _, p := range d.Partitions {
			if p.MountPoint != "" {
				mp[d.Device] = append(mp[d.Device], p.MountPoint)
			}
		}
	}
	return mp
}

func wrapExecErr(what string, err error) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return fmt.Errorf("%s failed: %s", what, strings.TrimSpace(string(ee.Stderr)))
	}
	return fmt.Errorf("%s error: %w", what, err)
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func preferDevPath(deviceNode, bsd string) string {
	if strings.HasPrefix(deviceNode, "/dev/") {
		return deviceNode
	}
	if bsd != "" {
		return filepath.Join("/dev", bsd)
	}
	return deviceNode
}
