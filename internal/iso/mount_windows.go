//go:build windows

package iso

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type windowsMounter struct{}

func platformMounter() Mounter { return &windowsMounter{} }

func (m *windowsMounter) IsSupported() bool { return true }

func (m *windowsMounter) Mount(ctx context.Context, isoPath string) (*MountResult, error) {
	// Ensure timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// PowerShell script to mount ISO and return drive letter
	// Use single-quoted string to prevent command injection
	script := fmt.Sprintf(`
$ErrorActionPreference = "Stop"
$img = Mount-DiskImage -ImagePath '%s' -PassThru
$vol = $img | Get-Volume
if (-not $vol.DriveLetter) {
    throw "ISO mounted but no drive letter assigned"
}
$vol.DriveLetter
`, escapePSString(isoPath))

	out, err := m.runPS(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("mount ISO %s: %w", isoPath, err)
	}

	letter := strings.TrimSpace(string(out))
	if letter == "" {
		return nil, fmt.Errorf("mount ISO %s: no drive letter returned", isoPath)
	}

	return &MountResult{
		MountPath: letter + `:\`,
		DevEntry:  isoPath, // Use ISO path for dismounting
	}, nil
}

func (m *windowsMounter) Unmount(ctx context.Context, result *MountResult) error {
	if result == nil {
		return nil
	}

	isoPath := result.DevEntry
	if isoPath == "" {
		return nil
	}

	unmountCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	script := fmt.Sprintf(`Dismount-DiskImage -ImagePath '%s' -ErrorAction SilentlyContinue`, escapePSString(isoPath))
	m.runPS(unmountCtx, script) // Best-effort, ignore errors
	return nil
}

func (m *windowsMounter) DetachExisting(ctx context.Context, isoPath string) string {
	escaped := escapePSString(isoPath)
	script := fmt.Sprintf(`
$img = Get-DiskImage -ImagePath '%s' -ErrorAction SilentlyContinue
if ($img -and $img.Attached) {
    Dismount-DiskImage -ImagePath '%s' -ErrorAction SilentlyContinue | Out-Null
    "detached"
}
`, escaped, escaped)

	out, err := m.runPS(ctx, script)
	if err != nil {
		return ""
	}

	if strings.TrimSpace(string(out)) == "detached" {
		return isoPath
	}
	return ""
}

// runPS executes a PowerShell script and returns stdout.
func (m *windowsMounter) runPS(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-Command", script,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, wrapExecErr("powershell", err, out)
	}
	return out, nil
}

func wrapExecErr(what string, err error, output []byte) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		stderr := strings.TrimSpace(string(ee.Stderr))
		if stderr == "" && len(output) > 0 {
			stderr = strings.TrimSpace(string(output))
		}
		if stderr != "" {
			return fmt.Errorf("%s failed: %s", what, stderr)
		}
	}
	return fmt.Errorf("%s error: %w", what, err)
}

// escapePSString escapes a string for use in a PowerShell single-quoted literal.
// Single quotes are the only characters that need escaping (doubled).
func escapePSString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
