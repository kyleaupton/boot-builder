//go:build linux

package priv

import (
	"context"
	"errors"
)

type linuxRunner struct{}

func platformRunner() Runner { return &linuxRunner{} }

func (l *linuxRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return "", errors.New("privileged runner not implemented on linux yet")
}
