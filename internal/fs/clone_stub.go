//go:build !darwin

package fs

import (
	"io"
	"os"
	"path/filepath"
)

// TempISODir is where cloned ISOs are stored
var TempISODir = filepath.Join(os.TempDir(), "flashit")

// CloneFile copies src to dst (no APFS clone on non-darwin)
func CloneFile(src, dst string) error {
	// Ensure temp directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

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
