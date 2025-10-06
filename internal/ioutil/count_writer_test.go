package ioutil

import (
	"errors"
	"io"
	"testing"
)

type mockWriter struct {
	written  []byte
	writeErr error
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.written = append(m.written, p...)
	return len(p), nil
}

func TestNewCountWriter(t *testing.T) {
	w := &mockWriter{}
	cw := NewCountWriter(w)

	if cw.w != w {
		t.Error("underlying writer not set correctly")
	}

	if cw.count != 0 {
		t.Errorf("initial count should be 0, got %d", cw.count)
	}
}

func TestCountWriter_Write(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		writeErr  error
		expectN   int
		expectErr error
	}{
		{
			name:      "normal write",
			data:      []byte("test data"),
			expectN:   9,
			expectErr: nil,
		},
		{
			name:      "empty write",
			data:      []byte{},
			expectN:   0,
			expectErr: nil,
		},
		{
			name:      "write error",
			data:      []byte("test data"),
			writeErr:  io.ErrShortWrite,
			expectN:   0,
			expectErr: io.ErrShortWrite,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := &mockWriter{writeErr: tc.writeErr}
			cw := NewCountWriter(w)

			n, err := cw.Write(tc.data)

			if !errors.Is(err, tc.expectErr) {
				t.Errorf("expected error %v, got %v", tc.expectErr, err)
			}

			if n != tc.expectN {
				t.Errorf("expected to write %d bytes, got %d", tc.expectN, n)
			}

			if err == nil {
				if cw.count != int64(tc.expectN) {
					t.Errorf("expected count to be %d, got %d", tc.expectN, cw.count)
				}
			}
		})
	}
}

func TestCountWriter_Count(t *testing.T) {
	w := &mockWriter{}
	cw := NewCountWriter(w)

	data1 := []byte("first write")
	data2 := []byte("second write")

	_, _ = cw.Write(data1)
	if count := cw.Count(); count != int64(len(data1)) {
		t.Errorf("expected count %d, got %d", len(data1), count)
	}

	_, _ = cw.Write(data2)
	expectedTotal := int64(len(data1) + len(data2))
	if count := cw.Count(); count != expectedTotal {
		t.Errorf("expected count %d, got %d", expectedTotal, count)
	}
}

func TestCountWriter_MultipleWrites(t *testing.T) {
	w := &mockWriter{}
	cw := NewCountWriter(w)

	data := [][]byte{
		[]byte("first chunk"),
		[]byte("second chunk"),
		[]byte("third chunk"),
	}

	expectedTotal := 0
	for _, chunk := range data {
		n, err := cw.Write(chunk)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedTotal += n
	}

	if count := cw.Count(); count != int64(expectedTotal) {
		t.Errorf("expected total count %d, got %d", expectedTotal, count)
	}
}
