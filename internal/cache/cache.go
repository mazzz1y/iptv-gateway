package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iptv-gateway/internal/ctxutil"
	"iptv-gateway/internal/logging"
	"iptv-gateway/internal/metrics"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	compressedExtension   = ".gz"
	uncompressedExtension = ".cache"
	metaExtension         = ".meta"
)

type Cache struct {
	directHttpClient *http.Client
	dir              string
	cleanupTicker    *time.Ticker
	doneCh           chan struct{}
	ttl              time.Duration
	retention        time.Duration
	compression      bool
}

func NewCache(cacheDir string, ttl, retention time.Duration, compression bool) (*Cache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &Cache{
		dir:              cacheDir,
		cleanupTicker:    time.NewTicker(24 * time.Hour),
		doneCh:           make(chan struct{}),
		directHttpClient: newDirectHTTPClient(),
		ttl:              ttl,
		retention:        retention,
		compression:      compression,
	}

	if retention > 0 {
		go cache.cleanupRoutine()
	}

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
	fileExt := c.fileExt()

	reader := &Reader{
		URL:         url,
		Name:        name,
		FilePath:    filepath.Join(c.dir, name+fileExt),
		MetaPath:    filepath.Join(c.dir, name+metaExtension),
		client:      c.directHttpClient,
		ttl:         c.ttl,
		compression: c.compression,
	}

	var err error
	var readCloser io.ReadCloser
	var cacheStatus = metrics.CacheStatusMiss

	s := reader.checkCacheStatus()

	switch s {
	case statusValid, statusRenewed:
		readCloser, err = reader.newCachedReader()
		if err == nil {
			if s == statusValid {
				cacheStatus = metrics.CacheStatusHit
			} else {
				cacheStatus = metrics.CacheStatusRenewed
			}
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

	logging.Debug(
		ctx, "file access", "cache", formatCacheStatus(s), "url", logging.SanitizeURL(url))

	metrics.ProxyRequestsTotal.WithLabelValues(
		ctxutil.ClientName(ctx),
		ctxutil.RequestType(ctx),
		cacheStatus,
	).Inc()

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

func (c *Cache) fileExt() string {
	if c.compression {
		return compressedExtension
	}
	return uncompressedExtension
}

func (c *Cache) removeEntry(name string) error {
	dataPath := filepath.Join(c.dir, name+c.fileExt())
	metaPath := filepath.Join(c.dir, name+".meta")

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

func (c *Cache) cleanupRoutine() {
	for {
		select {
		case <-c.cleanupTicker.C:
			if err := c.cleanExpired(); err != nil {
				logging.Error(context.Background(), err, "failed to clean expired Cache")
				return
			}
		case <-c.doneCh:
			return
		}
	}
}

func (c *Cache) cleanExpired() error {
	allFiles, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to list cache directory: %w", err)
	}

	expiredRemoved := 0
	orphanedRemoved := 0
	now := time.Now()

	expectedExt := c.fileExt()

	for _, file := range allFiles {
		fileName := file.Name()
		filePath := filepath.Join(c.dir, fileName)

		switch {
		case strings.HasSuffix(fileName, expectedExt):
			name := strings.TrimSuffix(fileName, expectedExt)
			metaPath := filepath.Join(c.dir, name+".meta")

			_, err := os.Stat(metaPath)
			if os.IsNotExist(err) {
				if err := os.Remove(filePath); err != nil {
					return fmt.Errorf("failed to remove orphaned file: %w", err)
				}
				orphanedRemoved++
			}

		case strings.HasSuffix(fileName, ".meta"):
			name := strings.TrimSuffix(fileName, ".meta")

			isExpired := false
			metaPath := filepath.Join(c.dir, fileName)
			metadata, err := readMetadata(metaPath)
			if err == nil && now.Sub(time.Unix(metadata.CachedAt, 0)) > c.retention {
				isExpired = true
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
