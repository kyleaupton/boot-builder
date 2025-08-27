package steps

import (
	"boot-builder/internal/core"
	"boot-builder/internal/fs"
	"boot-builder/internal/priv"
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// DarwinUnmountDisk unmounts all volumes on a disk (e.g. /dev/disk4)
type DarwinUnmountDisk struct{ Device string }

func (d DarwinUnmountDisk) Name() string            { return "Unmount target" }
func (d DarwinUnmountDisk) Estimate() time.Duration { return 2 * time.Second }
func (d DarwinUnmountDisk) Run(ctx context.Context, e core.Executor) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	if d.Device == "" {
		return errors.New("device not specified")
	}
	out, err := priv.Run(ctx, "/usr/sbin/diskutil", "unmountDisk", "force", d.Device)
	if err != nil {
		// Non-fatal if already unmounted; still report
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("diskutil unmountDisk: %s", strings.TrimSpace(out))})
		return nil
	}
	e.Emit(core.Event{Type: "log", Message: strings.TrimSpace(out)})
	return nil
}

// DarwinRawWriteISO writes the ISO byte-for-byte to the raw disk (e.g. /dev/rdisk4)
type DarwinRawWriteISO struct{ ISOPath, Device string }

func (d DarwinRawWriteISO) Name() string            { return "Write image" }
func (d DarwinRawWriteISO) Estimate() time.Duration { return 60 * time.Second }
func (d DarwinRawWriteISO) Run(ctx context.Context, e core.Executor) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	if d.ISOPath == "" || d.Device == "" {
		return errors.New("missing ISOPath or Device")
	}

	// Prefer raw device node for speed
	dev := d.Device
	if strings.HasPrefix(dev, "/dev/disk") {
		dev = strings.Replace(dev, "/dev/disk", "/dev/rdisk", 1)
	}
	progress := func(wrote, total int64) {
		if total > 0 {
			e.Emit(core.Event{Type: "progress", Percent: (float64(wrote) / float64(total)) * 100})
		}
	}
	return fs.RawWriteToBlockDevice(ctx, d.ISOPath, dev, progress)
}

// DarwinEjectDisk ejects the disk
type DarwinEjectDisk struct{ Device string }

func (d DarwinEjectDisk) Name() string            { return "Eject" }
func (d DarwinEjectDisk) Estimate() time.Duration { return 2 * time.Second }
func (d DarwinEjectDisk) Run(ctx context.Context, e core.Executor) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	if d.Device == "" {
		return errors.New("device not specified")
	}
	out, err := priv.Run(ctx, "/usr/sbin/diskutil", "eject", d.Device)
	if err != nil {
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("diskutil eject: %s", strings.TrimSpace(out))})
		return nil
	}
	e.Emit(core.Event{Type: "log", Message: strings.TrimSpace(out)})
	return nil
}
