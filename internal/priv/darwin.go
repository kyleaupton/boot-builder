//go:build darwin

package priv

import (
	"context"
	"os/exec"
	"strings"
)

type darwinRunner struct{}

func platformRunner() Runner { return &darwinRunner{} }

func (d *darwinRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	// Short-term: assume the process is already elevated (run app/CLI with sudo)
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
