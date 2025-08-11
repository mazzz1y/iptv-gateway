package ioutil

import "io"

type ReaderWithCloser struct {
	reader io.Reader
	closer func() error
}

func NewReaderWithCloser(reader io.Reader, closer func() error) *ReaderWithCloser {
	return &ReaderWithCloser{reader: reader, closer: closer}
}

func (w *ReaderWithCloser) Read(p []byte) (n int, err error) {
	if w.reader == nil {
		return 0, io.EOF
	}
	return w.reader.Read(p)
}

func (w *ReaderWithCloser) Close() error {
	if w.closer != nil {
		return w.closer()
	}
	return nil
}
