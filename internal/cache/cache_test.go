package cache

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanExpired(t *testing.T) {
	tests := []struct {
		name           string
		ttl            time.Duration
		expectValid    bool
		expectExpired  bool
		expectOrphaned bool
	}{
		{
			name:           "clean expired and orphaned files",
			ttl:            time.Hour,
			expectValid:    true,
			expectExpired:  false,
			expectOrphaned: false,
		},
		{
			name:           "zero ttl keeps expired files",
			ttl:            0,
			expectValid:    true,
			expectExpired:  true,
			expectOrphaned: false,
		},
		{
			name:           "negative ttl keeps expired files",
			ttl:            -1 * time.Hour,
			expectValid:    true,
			expectExpired:  true,
			expectOrphaned: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			validName := "valid"
			validFile := filepath.Join(tmpDir, validName+fileExtension)
			validMeta := filepath.Join(tmpDir, validName+metaExtension)
			if err := os.WriteFile(validFile, []byte("valid content"), 0644); err != nil {
				t.Fatalf("failed to create valid file: %v", err)
			}
			if err := createTestMetadata(validMeta, time.Now().Unix()); err != nil {
				t.Fatalf("failed to create valid metadata: %v", err)
			}

			expiredName := "expired"
			expiredFile := filepath.Join(tmpDir, expiredName+fileExtension)
			expiredMeta := filepath.Join(tmpDir, expiredName+metaExtension)
			if err := os.WriteFile(expiredFile, []byte("expired content"), 0644); err != nil {
				t.Fatalf("failed to create expired file: %v", err)
			}
			if err := createTestMetadata(expiredMeta, time.Now().Add(-2*time.Hour).Unix()); err != nil {
				t.Fatalf("failed to create expired metadata: %v", err)
			}

			orphanedName := "orphaned"
			orphanedFile := filepath.Join(tmpDir, orphanedName+fileExtension)
			if err := os.WriteFile(orphanedFile, []byte("orphaned content"), 0644); err != nil {
				t.Fatalf("failed to create orphaned file: %v", err)
			}

			cache := &Cache{
				dir: tmpDir,
			}
			if err := cache.cleanExpired(tmpDir, tc.ttl); err != nil {
				t.Fatalf("failed to clean cache: %v", err)
			}

			checkFileExists(t, validFile, tc.expectValid)
			checkFileExists(t, validMeta, tc.expectValid)
			checkFileExists(t, expiredFile, tc.expectExpired)
			checkFileExists(t, expiredMeta, tc.expectExpired)
			checkFileExists(t, orphanedFile, tc.expectOrphaned)
		})
	}
}

func TestCache_NewReader(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &Cache{
		dir:        tmpDir,
		ttl:        time.Hour,
		httpClient: &http.Client{},
	}

	if cache.dir != tmpDir {
		t.Errorf("expected cache directory to be %s, got %s", tmpDir, cache.dir)
	}

	if cache.ttl != time.Hour {
		t.Errorf("expected ttl to be %s, got %s", time.Hour, cache.ttl)
	}

	if cache.httpClient == nil {
		t.Error("client should not be nil")
	}
}

func createTestMetadata(path string, cachedAt int64) error {
	metadata := Metadata{
		LastModified: time.Now().Add(-2 * time.Hour).Unix(),
		CachedAt:     cachedAt,
		ContentType:  "text/plain",
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(metadata)
}

func checkFileExists(t *testing.T, path string, shouldExist bool) {
	_, err := os.Stat(path)
	if shouldExist && os.IsNotExist(err) {
		t.Errorf("file %s should exist but doesn't", path)
	} else if !shouldExist && !os.IsNotExist(err) {
		t.Errorf("file %s should not exist but does", path)
	}
}
