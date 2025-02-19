package compressed

import (
	"compress/gzip"
	"io"
)

func Write(src io.Reader, dst io.Writer) (written int64, err error) {
	gzw, err := gzip.NewWriterLevel(dst, gzip.BestCompression)
	if err != nil {
		return
	}
	written, err = io.Copy(gzw, src)
	if err != nil {
		return
	}
	err = gzw.Close()
	return
}

func Read(src io.Reader, dst io.Writer) (written int64, err error) {
	gzr, err := gzip.NewReader(src)
	if err != nil {
		return
	}
	written, err = io.Copy(dst, gzr)
	if err != nil {
		return
	}
	err = gzr.Close()
	return
}
