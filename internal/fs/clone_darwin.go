//go:build darwin

package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// TempISODir is where cloned ISOs are stored to bypass TCC restrictions
const TempISODir = "/private/var/tmp/flashit"

// CloneFile creates an APFS clone of src at dst.
// Falls back to regular copy if cloning is not supported.
func CloneFile(src, dst string) error {
	// Ensure temp directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	// Try APFS clone first (instant for large files)
	err := unix.Clonefile(src, dst, 0)
	if err == nil {
		return nil
	}

	// Fall back to regular copy if clone not supported
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
