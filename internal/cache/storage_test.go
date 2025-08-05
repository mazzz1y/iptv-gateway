package cache

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "test"+metaExtension)

	cachedTime := time.Now().Add(-time.Hour).Unix()
	metadata := Metadata{
		LastModified: time.Now().Add(-2 * time.Hour).Unix(),
		CachedAt:     cachedTime,
		ContentType:  "text/plain",
	}

	file, err := os.Create(metaPath)
	if err != nil {
		t.Fatalf("failed to create test metadata file: %v", err)
	}

	if err := json.NewEncoder(file).Encode(metadata); err != nil {
		file.Close()
		t.Fatalf("failed to write test metadata: %v", err)
	}
	file.Close()

	readMetadata, err := ReadMetadata(metaPath)
	if err != nil {
		t.Fatalf("failed to read metadata: %v", err)
	}

	if readMetadata.CachedAt != cachedTime {
		t.Errorf("cachedAt time mismatch, got: %d, want: %d", readMetadata.CachedAt, cachedTime)
	}

	if readMetadata.ContentType != metadata.ContentType {
		t.Errorf("contentType mismatch, got: %s, want: %s", readMetadata.ContentType, metadata.ContentType)
	}
}

func TestReadMetadata_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "invalid"+metaExtension)

	if err := os.WriteFile(metaPath, []byte("invalid metadata format"), 0644); err != nil {
		t.Fatalf("failed to create invalid metadata: %v", err)
	}

	_, err := ReadMetadata(metaPath)
	if err == nil {
		t.Error("Expected error for invalid metadata format, got nil")
	}
}

func TestSaveMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test"+fileExtension)
	metaPath := filepath.Join(tmpDir, "test"+metaExtension)

	contentType := "text/plain"
	lastModified := time.Now().Format(time.RFC1123)

	header := http.Header{}
	header.Set("Content-Type", contentType)
	header.Set("Last-Modified", lastModified)

	resp := &http.Response{
		Header: header,
		Body:   http.NoBody,
	}

	reader := &Reader{
		FilePath:    filePath,
		MetaPath:    metaPath,
		res:         resp,
		contentType: contentType,
	}

	err := reader.SaveMetadata()
	if err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	metadata, err := ReadMetadata(metaPath)
	if err != nil {
		t.Fatalf("failed to read saved metadata: %v", err)
	}

	if metadata.ContentType != contentType {
		t.Errorf("Content type mismatch, got: %s, want: %s", metadata.ContentType, contentType)
	}

	if metadata.CachedAt <= 0 {
		t.Error("cachedAt time should be positive")
	}

	lastModifiedTime, _ := time.Parse(time.RFC1123, lastModified)
	if metadata.LastModified != lastModifiedTime.Unix() {
		t.Errorf("lastModified time mismatch, got: %d, want: %d", metadata.LastModified, lastModifiedTime.Unix())
	}
}
