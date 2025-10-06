package cache

import (
	"context"
	"encoding/json"
	"io"
	"majmun/internal/config"
	"majmun/internal/config/common"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	t.Run("successfully creates cache", func(t *testing.T) {
		tmpDir := t.TempDir()
		ttl := time.Hour
		retention := 24 * time.Hour

		cache, err := NewCache(config.CacheConfig{
			Path:        tmpDir,
			TTL:         common.Duration(ttl),
			Retention:   common.Duration(retention),
			Compression: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if cache.dir != tmpDir {
			t.Errorf("expected dir %s, got %s", tmpDir, cache.dir)
		}

		if cache.directHttpClient == nil {
			t.Error("expected httpClient to be set")
		}

		if cache.cleanupTicker == nil {
			t.Error("expected cleanupTicker to be set")
		}

		if cache.doneCh == nil {
			t.Error("expected doneCh to be set")
		}

		cache.Close()
	})

	t.Run("creates cache directory", func(t *testing.T) {
		tmpDir := filepath.Join(t.TempDir(), "nested", "cache")

		cache, err := NewCache(config.CacheConfig{
			Path:        tmpDir,
			TTL:         common.Duration(time.Hour),
			Retention:   common.Duration(24 * time.Hour),
			Compression: true,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer cache.Close()

		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
			t.Error("expected cache directory to be created")
		}
	})

	t.Run("fails with invalid directory", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("skipping test when running as root")
		}

		tmpFile := filepath.Join(t.TempDir(), "blocking-file")
		if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create blocking file: %v", err)
		}

		invalidDir := filepath.Join(tmpFile, "cache")
		_, err := NewCache(config.CacheConfig{
			Path:        invalidDir,
			TTL:         common.Duration(time.Hour),
			Retention:   common.Duration(24 * time.Hour),
			Compression: true,
		})
		if err == nil {
			t.Error("expected error for invalid directory")
		}
	})
}

func TestCache_NewCachedHTTPClient(t *testing.T) {
	cache, err := NewCache(config.CacheConfig{
		Path:        t.TempDir(),
		TTL:         common.Duration(time.Hour),
		Retention:   common.Duration(24 * time.Hour),
		Compression: true,
	})
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	client := cache.NewCachedHTTPClient()

	if client.Timeout != 10*time.Minute {
		t.Errorf("expected timeout 10m, got %v", client.Timeout)
	}

	if client.Transport == nil {
		t.Error("expected transport to be set")
	}

	t.Run("redirect limit", func(t *testing.T) {
		req := &http.Request{}
		via := make([]*http.Request, 5)

		err := client.CheckRedirect(req, via)
		if err == nil {
			t.Error("expected error for too many redirects")
		}
	})
}

func TestCache_NewReader(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewCache(config.CacheConfig{
		Path:        tmpDir,
		TTL:         common.Duration(time.Hour),
		Retention:   common.Duration(24 * time.Hour),
		Compression: true,
	})

	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	t.Run("creates reader with correct properties", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("#EXTM3U\n#EXTINF:10.0,\ntest.ts\n"))
		}))
		defer server.Close()

		reader, err := cache.NewReader(ctx, server.URL)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer func() { _ = reader.Close() }()

		if reader.URL != server.URL {
			t.Errorf("expected URL %s, got %s", server.URL, reader.URL)
		}

		if reader.Name == "" {
			t.Error("expected Name to be set")
		}

		if reader.FilePath == "" {
			t.Error("expected FilePath to be set")
		}

		if reader.MetaPath == "" {
			t.Error("expected MetaPath to be set")
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read content: %v", err)
		}

		expected := "#EXTM3U\n#EXTINF:10.0,\ntest.ts\n"
		if string(content) != expected {
			t.Errorf("expected content %q, got %q", expected, string(content))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := cache.NewReader(ctx, server.URL)
		if err == nil {
			t.Error("expected error for server error")
		}
	})
}

func TestCache_CleanExpired(t *testing.T) {
	tests := []struct {
		name           string
		ttl            time.Duration
		retention      time.Duration
		expectValid    bool
		expectExpired  bool
		expectOrphaned bool
	}{
		{
			name:           "clean expired and orphaned files",
			ttl:            time.Hour,
			retention:      time.Hour,
			expectValid:    true,
			expectExpired:  false,
			expectOrphaned: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cache, err := NewCache(config.CacheConfig{
				Path:        tmpDir,
				TTL:         common.Duration(tc.ttl),
				Retention:   common.Duration(tc.retention),
				Compression: true,
			})
			if err != nil {
				t.Fatalf("failed to create cache: %v", err)
			}
			defer cache.Close()

			validName := "valid"
			validFile := filepath.Join(tmpDir, validName+compressedExtension)
			validMeta := filepath.Join(tmpDir, validName+metaExtension)
			if err := os.WriteFile(validFile, []byte("valid content"), 0644); err != nil {
				t.Fatalf("failed to create valid file: %v", err)
			}
			if err := createTestMetadata(validMeta, time.Now().Unix()); err != nil {
				t.Fatalf("failed to create valid metadata: %v", err)
			}

			expiredName := "expired"
			expiredFile := filepath.Join(tmpDir, expiredName+compressedExtension)
			expiredMeta := filepath.Join(tmpDir, expiredName+metaExtension)
			if err := os.WriteFile(expiredFile, []byte("expired content"), 0644); err != nil {
				t.Fatalf("failed to create expired file: %v", err)
			}
			if err := createTestMetadata(expiredMeta, time.Now().Add(-2*time.Hour).Unix()); err != nil {
				t.Fatalf("failed to create expired metadata: %v", err)
			}

			orphanedName := "orphaned"
			orphanedFile := filepath.Join(tmpDir, orphanedName+compressedExtension)
			if err := os.WriteFile(orphanedFile, []byte("orphaned content"), 0644); err != nil {
				t.Fatalf("failed to create orphaned file: %v", err)
			}

			if err := cache.cleanExpired(); err != nil {
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

func TestCache_RemoveEntry(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewCache(config.CacheConfig{
		Path:        tmpDir,
		TTL:         common.Duration(time.Hour),
		Retention:   common.Duration(24 * time.Hour),
		Compression: true,
	})

	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer cache.Close()

	name := "test"
	dataPath := filepath.Join(tmpDir, name+compressedExtension)
	metaPath := filepath.Join(tmpDir, name+metaExtension)

	if err := os.WriteFile(dataPath, []byte("test data"), 0644); err != nil {
		t.Fatalf("failed to create data file: %v", err)
	}
	if err := createTestMetadata(metaPath, time.Now().Unix()); err != nil {
		t.Fatalf("failed to create metadata: %v", err)
	}

	if err := cache.removeEntry(name); err != nil {
		t.Fatalf("failed to remove entry: %v", err)
	}

	checkFileExists(t, dataPath, false)
	checkFileExists(t, metaPath, false)

	if err := cache.removeEntry("nonexistent"); err != nil {
		t.Errorf("unexpected error removing non-existent entry: %v", err)
	}
}

func TestNewDirectHTTPClient(t *testing.T) {
	client := newDirectHTTPClient(nil)

	if client.Timeout != 10*time.Minute {
		t.Errorf("expected timeout 10m, got %v", client.Timeout)
	}

	transport, ok := client.Transport.(*headerTransport)
	if !ok {
		t.Errorf("expected *headerTransport, got %T", client.Transport)
	} else if transport.base != http.DefaultTransport {
		t.Error("expected default transport to be wrapped")
	}

	req := &http.Request{}
	via := make([]*http.Request, 5)

	err := client.CheckRedirect(req, via)
	if err == nil {
		t.Error("expected error for too many redirects")
	}
}

func createTestMetadata(path string, cachedAt int64) error {
	metadata := Metadata{
		CachedAt: cachedAt,
		Headers:  make(map[string]string, len(forwardedHeaders)),
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	return json.NewEncoder(file).Encode(metadata)
}

func checkFileExists(t *testing.T, path string, shouldExist bool) {
	t.Helper()
	_, err := os.Stat(path)
	if shouldExist && os.IsNotExist(err) {
		t.Errorf("file %s should exist but doesn't", path)
	} else if !shouldExist && !os.IsNotExist(err) {
		t.Errorf("file %s should not exist but does", path)
	}
}
