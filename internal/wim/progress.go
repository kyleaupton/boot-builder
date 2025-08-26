//go:build (darwin || linux) && wimlib

package wim

/*
#include <wimlib.h>
*/
import "C"

type ProgressPhase int

const (
	PhaseUnknown ProgressPhase = iota
	PhaseScanning
	PhaseWritingStreams
	PhaseWritingMetadata
	PhaseVerifying
	PhaseSplittingBegin
	PhaseSplittingEnd
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

func mapPhase(msg int, hint int) ProgressPhase {
	switch msg {
	case int(C.WIMLIB_PROGRESS_MSG_WRITE_STREAMS):
		return PhaseWritingStreams
	case int(C.WIMLIB_PROGRESS_MSG_VERIFY_INTEGRITY):
		return PhaseVerifying
	default:
		if hint == 2 {
			return PhaseWritingStreams
		}
		if hint == 4 {
			return PhaseVerifying
		}
		return PhaseUnknown
	}
}
