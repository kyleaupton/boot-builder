package iso

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
)

// ISO 9660 constants
const (
	// Primary Volume Descriptor starts at sector 16 (2048 bytes per sector)
	pvdOffset = 32768

	// "CD001" signature is at byte 1 of the PVD (byte 0 is the descriptor type)
	iso9660SignatureOffset = pvdOffset + 1
	iso9660Signature       = "CD001"

	// Volume label is at offset 40 in the PVD (32 bytes, space-padded)
	volumeLabelOffset = pvdOffset + 40
	volumeLabelLength = 32

	// MBR boot signature at bytes 510-511
	mbrSignatureOffset = 510
)

var (
	// ErrNotISO9660 indicates the file lacks the ISO 9660 "CD001" signature.
	ErrNotISO9660 = errors.New("file is not a valid ISO 9660 image")

	// ErrNotHybrid indicates the ISO lacks an MBR boot signature,
	// meaning it cannot be written directly to a USB drive.
	ErrNotHybrid = errors.New("ISO cannot be written directly to USB (non-hybrid)")

	// ErrFileTooSmall indicates the file is too small to contain required headers.
	ErrFileTooSmall = errors.New("file too small to be a valid ISO")
)

// ISOInfo contains metadata extracted from an ISO image.
type ISOInfo struct {
	Path        string // Original file path
	VolumeLabel string // Volume label from PVD (trimmed)
	Size        int64  // File size in bytes
	IsISO9660   bool   // Has valid CD001 signature
	IsHybrid    bool   // Has MBR boot signature (0x55AA)
}

// Inspect reads an ISO file and extracts metadata including hybrid detection.
func Inspect(path string) (*ISOInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	info := &ISOInfo{
		Path: path,
		Size: stat.Size(),
	}

	// Minimum size: need at least PVD offset + 1 (type byte) + signature length
	minSize := int64(iso9660SignatureOffset + len(iso9660Signature))
	if stat.Size() < minSize {
		return info, ErrFileTooSmall
	}

	// Check ISO 9660 signature at PVD offset + 1 (byte 0 is descriptor type)
	sigBuf := make([]byte, len(iso9660Signature))
	if _, err := f.ReadAt(sigBuf, iso9660SignatureOffset); err != nil {
		return info, err
	}
	info.IsISO9660 = string(sigBuf) == iso9660Signature

	if !info.IsISO9660 {
		return info, nil
	}

	// Extract volume label (32 bytes at PVD+40)
	labelBuf := make([]byte, volumeLabelLength)
	if _, err := f.ReadAt(labelBuf, volumeLabelOffset); err != nil && err != io.EOF {
		return info, err
	}
	info.VolumeLabel = strings.TrimRight(string(labelBuf), " ")

	// Check MBR boot signature (0x55 0xAA at offset 510-511)
	mbrBuf := make([]byte, 2)
	if _, err := f.ReadAt(mbrBuf, mbrSignatureOffset); err != nil {
		return info, err
	}
	info.IsHybrid = bytes.Equal(mbrBuf, []byte{0x55, 0xAA})

	return info, nil
}

// ValidateHybrid is a quick check for Plan() to verify an ISO can be
// written directly to USB. Returns nil if valid, or an appropriate error.
func ValidateHybrid(path string) error {
	info, err := Inspect(path)
	if err != nil {
		// ErrFileTooSmall is still a valid inspection result, check IsISO9660
		if !errors.Is(err, ErrFileTooSmall) {
			return err
		}
		return ErrFileTooSmall
	}

	if !info.IsISO9660 {
		return ErrNotISO9660
	}

	if !info.IsHybrid {
		return ErrNotHybrid
	}

	return nil
}
