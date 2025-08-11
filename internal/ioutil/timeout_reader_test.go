package ioutil

import (
	"errors"
	"io"
	"testing"
	"time"
)

type mockReader struct {
	delay    time.Duration
	response []byte
	err      error
}

func (m *mockReader) Read(p []byte) (int, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if len(m.response) > 0 {
		n := copy(p, m.response)
		return n, m.err
	}

	return 0, m.err
}

func TestTimeoutReader_Read_Success(t *testing.T) {
	expectedData := []byte("test data")
	mock := &mockReader{
		delay:    10 * time.Millisecond,
		response: expectedData,
		err:      nil,
	}

	reader := NewTimeoutReader(mock, 100*time.Millisecond)
	buf := make([]byte, 100)

	n, err := reader.Read(buf)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if n != len(expectedData) {
		t.Fatalf("expected to read %d bytes, got: %d", len(expectedData), n)
	}

	if string(buf[:n]) != string(expectedData) {
		t.Fatalf("expected data '%s', got: '%s'", expectedData, buf[:n])
	}
}

func TestTimeoutReader_Read_Timeout(t *testing.T) {
	mock := &mockReader{
		delay: 100 * time.Millisecond,
	}

	reader := NewTimeoutReader(mock, 10*time.Millisecond)
	buf := make([]byte, 100)

	_, err := reader.Read(buf)

	if !errors.Is(err, ErrReaderTimeout) {
		t.Fatalf("expected errReaderTimeout, got: %v", err)
	}
}

func TestTimeoutReader_Read_UnderlyingError(t *testing.T) {
	expectedErr := errors.New("underlying error")
	mock := &mockReader{
		delay:    10 * time.Millisecond,
		response: []byte{},
		err:      expectedErr,
	}

	reader := NewTimeoutReader(mock, 100*time.Millisecond)
	buf := make([]byte, 100)

	_, err := reader.Read(buf)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("Expected error %v, got: %v", expectedErr, err)
	}
}

func TestTimeoutReader_Read_EOF(t *testing.T) {
	mock := &mockReader{
		delay:    10 * time.Millisecond,
		response: []byte{},
		err:      io.EOF,
	}

	reader := NewTimeoutReader(mock, 100*time.Millisecond)
	buf := make([]byte, 100)

	_, err := reader.Read(buf)

	if !errors.Is(err, io.EOF) {
		t.Fatalf("Expected EOF error, got: %v", err)
	}
}
