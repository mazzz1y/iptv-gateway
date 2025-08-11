package ioutil

import "io"

type CountWriter struct {
	w     io.Writer
	count int64
}

func NewCountWriter(w io.Writer) *CountWriter {
	return &CountWriter{
		w:     w,
		count: 0,
	}
}
func (cw *CountWriter) Write(p []byte) (n int, err error) {
	n, err = cw.w.Write(p)
	cw.count += int64(n)
	return n, err
}

func (cw *CountWriter) Count() int64 {
	return cw.count
}
