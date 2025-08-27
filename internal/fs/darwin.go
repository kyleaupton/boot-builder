//go:build darwin

package fs

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
)

type darwinFS struct{}

func platformFS() FileOps { return &darwinFS{} }

func (d *darwinFS) RawWriteToBlockDevice(ctx context.Context, srcPath string, devicePath string, onProgress ProgressFunc) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	dev := devicePath
	if strings.HasPrefix(dev, "/dev/disk") {
		dev = strings.Replace(dev, "/dev/disk", "/dev/rdisk", 1)
	}
	out, err := os.OpenFile(dev, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 4<<20)
	reader := bufio.NewReader(in)
	var written int64
	fi, _ := in.Stat()
	total := fi.Size()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, rerr := reader.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(written, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	return out.Sync()
}
