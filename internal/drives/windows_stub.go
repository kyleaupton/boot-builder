//go:build windows

package drives

import (
	"context"
	"errors"
)

type windowsProvider struct{}

func platformProvider() Provider { return &windowsProvider{} }

func (p *windowsProvider) ListRemovable(ctx context.Context) ([]Drive, error) {
	return nil, errors.New("drives: windows implementation not yet provided")
}

