package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iptv-gateway/internal/logging"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cache struct {
	httpClient    *http.Client
	dir           string
	ttl           time.Duration
	cleanupTicker *time.Ticker
	doneCh        chan struct{}
}

func NewCache(cacheDir string, ttl time.Duration) (*Cache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		dir:           cacheDir,
		ttl:           ttl,
		cleanupTicker: time.NewTicker(24 * time.Hour),
		doneCh:        make(chan struct{}),
		httpClient:    newDirectHTTPClient(),
	}

	go cache.cleanupRoutine(cacheDir, ttl)

	return cache, nil
}

func (c *Cache) NewCachedHTTPClient() *http.Client {
	return &http.Client{
		Transport: &cachingTransport{cache: c},
		Timeout:   10 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

func (c *Cache) NewReader(ctx context.Context, url string) (*Reader, error) {
	hash := sha256.Sum256([]byte(url))
	name := hex.EncodeToString(hash[:16])
	reader := &Reader{
		URL:      url,
		Name:     name,
		FilePath: filepath.Join(c.dir, name+fileExtension),
		MetaPath: filepath.Join(c.dir, name+metaExtension),
		client:   c.httpClient,
		ttl:      c.ttl,
	}

	s := reader.checkCacheStatus()

	var err error
	var readCloser io.ReadCloser

	switch s {
	case statusValid:
		readCloser, err = reader.newCachedReader()
		if err == nil {
			reader.ReadCloser = readCloser
		}

	case statusExpired, statusNotFound:
		readCloser, err = reader.newCachingReader(ctx)
		if err == nil {
			reader.ReadCloser = readCloser
		}

	default:
		readCloser, err = reader.newDirectReader(ctx)
		if err == nil {
			reader.ReadCloser = readCloser
		}
	}

	logging.Debug(ctx, "file access", "cache", formatCacheStatus(s), "url", logging.SanitizeURL(url))

	return reader, err
}

func (c *Cache) Close() {
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
	}

	if c.doneCh != nil {
		close(c.doneCh)
	}
}

func newDirectHTTPClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   10 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

func (c *Cache) removeEntry(name string) error {
	dataPath := filepath.Join(c.dir, name+fileExtension)
	metaPath := filepath.Join(c.dir, name+metaExtension)

	dataErr := os.Remove(dataPath)
	if dataErr != nil && !os.IsNotExist(dataErr) {
		return dataErr
	}

	metaErr := os.Remove(metaPath)
	if metaErr != nil && !os.IsNotExist(metaErr) {
		return metaErr
	}

	return nil
}

func (c *Cache) cleanExpired(cacheDir string, ttl time.Duration) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to access cache directory: %w", err)
	}

	allFiles, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to list cache directory: %w", err)
	}

	expiredRemoved := 0
	orphanedRemoved := 0
	now := time.Now()

	for _, file := range allFiles {
		fileName := file.Name()
		filePath := filepath.Join(cacheDir, fileName)

		switch {
		case strings.HasSuffix(fileName, fileExtension):
			name := strings.TrimSuffix(fileName, fileExtension)
			metaPath := filepath.Join(cacheDir, name+metaExtension)

			_, err := os.Stat(metaPath)
			if os.IsNotExist(err) {
				if err := os.Remove(filePath); err != nil {
					return fmt.Errorf("failed to remove orphaned file: %w", err)
				}
				orphanedRemoved++
			}

		case strings.HasSuffix(fileName, metaExtension):
			name := strings.TrimSuffix(fileName, metaExtension)

			isExpired := false
			if ttl > 0 {
				metaPath := filepath.Join(cacheDir, fileName)
				metadata, err := ReadMetadata(metaPath)
				if err == nil && now.Sub(time.Unix(metadata.CachedAt, 0)) > ttl {
					isExpired = true
				}
			}

			if isExpired {
				if err := c.removeEntry(name); err != nil {
					return err
				}
				expiredRemoved++
			}

		default:
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove unexpected file: %w", err)
			}
			orphanedRemoved++
		}
	}
	args := []any{"total_files", len(allFiles), "expired", expiredRemoved}
	logging.Info(context.Background(), "cache cleanup", args...)

	return nil
}

func (c *Cache) cleanupRoutine(cacheDir string, ttl time.Duration) {
	for {
		select {
		case <-c.cleanupTicker.C:
			if err := c.cleanExpired(cacheDir, ttl); err != nil {
				logging.Error(context.Background(), err, "failed to clean expired Cache")
				return
			}
		case <-c.doneCh:
			return
		}
	}
}
