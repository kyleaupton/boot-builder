//go:build !wimlib

package wim

import (
	"context"
	"errors"
)

type ProgressPhase int

const (
	PhaseUnknown ProgressPhase = iota
)

type Progress struct {
	Phase          ProgressPhase
	Percent        float64
	BytesCompleted uint64
	BytesTotal     uint64
	ItemsCompleted uint32
	ItemsTotal     uint32
	PartIndex      uint32
	PartCount      uint32
	Message        string
}

func SplitWithProgress(ctx context.Context, src, dstPrefix string, partSizeMiB int, check bool, cb func(Progress) bool) error {
	_ = ctx
	_ = src
	_ = dstPrefix
	_ = partSizeMiB
	_ = check
	_ = cb
	return errors.New("wimlib build tag not enabled; libwim not linked")
}
