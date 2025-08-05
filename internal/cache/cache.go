package cache

import (
	"context"
	"fmt"
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

	doneCh := make(chan struct{})

	cache := &Cache{
		httpClient:    NewHTTPClient(),
		dir:           cacheDir,
		ttl:           ttl,
		cleanupTicker: time.NewTicker(24 * time.Hour),
		doneCh:        doneCh,
	}

	go cache.cleanupRoutine(cacheDir, ttl)

	return cache, nil
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
	if orphanedRemoved > 0 {
		args = append(args, "orphaned", orphanedRemoved)
	}
	logging.Info(context.Background(), "cache cleanup", args...)

	return nil
}

func (c *Cache) Close() {
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
	}

	if c.doneCh != nil {
		close(c.doneCh)
	}
}

func (c *Cache) cleanupRoutine(cacheDir string, ttl time.Duration) {
	for {
		select {
		case <-c.cleanupTicker.C:
			if err := c.cleanExpired(cacheDir, ttl); err != nil {
				logging.Error(context.Background(), "failed to clean expired Cache", "error", err.Error())
				return
			}
		case <-c.doneCh:
			return
		}
	}
}
