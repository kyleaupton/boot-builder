package steps

import (
	"boot-builder/internal/core"
	"boot-builder/internal/fs"
	"boot-builder/internal/priv"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DarwinUnmountDisk unmounts all volumes on a disk (e.g. /dev/disk4)
// This does not require privileges for removable USB drives.
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
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "unmountDisk", "force", d.Device)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		// Check if it's already unmounted (not an error) or a real failure
		outLower := strings.ToLower(outStr)
		if strings.Contains(outLower, "not mounted") || strings.Contains(outLower, "already unmounted") {
			e.Emit(core.Event{Type: "log", Message: "Device already unmounted"})
			return nil
		}
		// Real error - fail the step
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("diskutil unmountDisk failed: %s", outStr)})
		return fmt.Errorf("failed to unmount %s: %w", d.Device, err)
	}
	e.Emit(core.Event{Type: "log", Message: outStr})
	return nil
}

// DarwinWriteLinuxISO writes a Linux ISO to disk using Apple's imaging tools.
// This is a single step that handles the entire pipeline:
// diskutil unmount → hdiutil convert → asr restore → eject
type DarwinWriteLinuxISO struct{ ISOPath, Device string }

func (d DarwinWriteLinuxISO) Name() string            { return "Write Linux ISO" }
func (d DarwinWriteLinuxISO) Estimate() time.Duration { return 5 * time.Minute } // Typically 3-10 minutes for 4GB ISO

func (d DarwinWriteLinuxISO) Run(ctx context.Context, e core.Executor) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	if d.ISOPath == "" || d.Device == "" {
		return errors.New("missing ISOPath or Device")
	}
	if d.Device == "/dev/disk0" || strings.HasSuffix(d.Device, "disk0") {
		return errors.New("refusing to write to /dev/disk0")
	}

	svc := priv.NewService()
	if err := svc.EnsureReady(ctx); err != nil {
		return err
	}

	// Clone ISO to neutral location to bypass TCC restrictions on ~/Downloads etc.
	// The privileged helper cannot read TCC-protected directories even as root.
	// APFS clone is near-instant for large files (metadata-only until COW divergence).
	tempPath := filepath.Join(fs.TempISODir, uuid.New().String()+".iso")
	e.Emit(core.Event{Type: "log", Message: "Preparing ISO..."})

	if err := fs.CloneFile(d.ISOPath, tempPath); err != nil {
		return fmt.Errorf("failed to prepare ISO: %w", err)
	}
	defer os.Remove(tempPath) // Cleanup on exit (success or failure)

	e.Emit(core.Event{Type: "log", Message: "Starting ISO write (this may take several minutes)..."})
	e.Emit(core.Event{Type: "log", Message: "Pipeline: unmount → convert → restore → eject"})

	// Progress callback to emit events during write
	progress := func(bytesWritten, totalBytes uint64) {
		if totalBytes > 0 {
			percent := float64(bytesWritten) * 100.0 / float64(totalBytes)
			e.Emit(core.Event{
				Type:    "progress",
				Percent: percent,
				Message: fmt.Sprintf("%.1f / %.1f GB", float64(bytesWritten)/1e9, float64(totalBytes)/1e9),
			})
		}
	}

	if err := svc.Disk().WriteISO(ctx, tempPath, d.Device, progress); err != nil {
		return fmt.Errorf("failed to write ISO: %w", err)
	}

	e.Emit(core.Event{Type: "log", Message: "ISO written successfully"})
	return nil
}

// DarwinEjectDisk ejects the disk
// This does not require privileges for removable USB drives.
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
	cmd := exec.CommandContext(ctx, "/usr/sbin/diskutil", "eject", d.Device)
	out, err := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		e.Emit(core.Event{Type: "log", Message: fmt.Sprintf("diskutil eject: %s", outStr)})
		return nil
	}
	e.Emit(core.Event{Type: "log", Message: outStr})
	return nil
}
