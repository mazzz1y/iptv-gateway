package cache

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCacheReader_BasicProperties(t *testing.T) {
	expectedContent := "test content"
	expectedContentType := "text/plain"

	reader := &Reader{
		ReadCloser:  io.NopCloser(strings.NewReader(expectedContent)),
		contentType: expectedContentType,
		client:      &http.Client{},
		ttl:         time.Hour,
	}

	if reader.GetContentType() != expectedContentType {
		t.Errorf("content type mismatch, got: %s, want: %s", reader.GetContentType(), expectedContentType)
	}

	if reader.ttl != time.Hour {
		t.Errorf("ttl mismatch, got: %s, want: %s", reader.ttl, time.Hour)
	}

	if reader.client == nil {
		t.Error("client should not be nil")
	}
}

func TestCacheReader_Close(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test"+fileExtension)
	metaPath := filepath.Join(tmpDir, "test"+metaExtension)

	readCloser := io.NopCloser(strings.NewReader("test"))
	gzipWriter := gzip.NewWriter(io.Discard)

	cr := &Reader{
		FilePath:   filePath,
		MetaPath:   metaPath,
		ReadCloser: readCloser,
		writer:     gzipWriter,
	}

	err := cr.Close()
	if err != nil {
		t.Fatalf("failed to close reader: %v", err)
	}

	err = cr.Close()
	if err != nil {
		t.Fatalf("failed to close reader a second time: %v", err)
	}
}

func TestCacheReader_FetchFromCache(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test"+fileExtension)
	metaPath := filepath.Join(tmpDir, "test"+metaExtension)

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	gw := gzip.NewWriter(f)
	gw.Write([]byte("test content"))
	gw.Close()
	f.Close()

	metadata := Metadata{
		ContentType: "text/plain",
		CachedAt:    time.Now().Unix(),
	}
	mf, _ := os.Create(metaPath)
	json.NewEncoder(mf).Encode(metadata)
	mf.Close()

	reader := &Reader{
		FilePath: filePath,
		MetaPath: metaPath,
	}

	readCloser, err := reader.newCachedReader()
	if err != nil {
		t.Fatalf("failed to fetch from cache: %v", err)
	}

	buf, err := io.ReadAll(readCloser)
	if err != nil {
		t.Fatalf("failed to read content: %v", err)
	}

	if string(buf) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(buf))
	}

	readCloser.Close()
}

func TestCacheReader_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test"+fileExtension)
	metaPath := filepath.Join(tmpDir, "test"+metaExtension)

	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(metaPath, []byte("test Metadata"), 0644); err != nil {
		t.Fatalf("failed to create test metadata: %v", err)
	}

	cr := &Reader{
		FilePath: filePath,
		MetaPath: metaPath,
	}

	cr.Cleanup()

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("file should be removed but still exists: %v", err)
	}

	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Errorf("metadata should be removed but still exists: %v", err)
	}
}

func TestCacheReader_VerifyDownload(t *testing.T) {
	tests := []struct {
		name          string
		expectedSize  int64
		bytesDownload int64
		expectResult  bool
	}{
		{
			name:          "download complete",
			expectedSize:  100,
			bytesDownload: 100,
			expectResult:  true,
		},
		{
			name:          "download incomplete",
			expectedSize:  100,
			bytesDownload: 50,
			expectResult:  false,
		},
		{
			name:          "unknown size",
			expectedSize:  0,
			bytesDownload: 100,
			expectResult:  true,
		},
		{
			name:          "negative size",
			expectedSize:  -1,
			bytesDownload: 100,
			expectResult:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cr := &Reader{
				expectedSize:  tc.expectedSize,
				bytesDownload: tc.bytesDownload,
			}

			result := cr.isDownloadComplete()

			if result != tc.expectResult {
				t.Errorf("expected %v, got %v", tc.expectResult, result)
			}
		})
	}
}
