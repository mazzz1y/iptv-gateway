package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var cachingHeaders = []string{
	"Cache-Control", "Expires", "Last-Modified", "ETag",
	"Age", "Vary", "Content-Type",
}

func ReadMetadata(metaPath string) (Metadata, error) {
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return Metadata{}, err
	}
	defer metaFile.Close()

	var m Metadata
	if err := json.NewDecoder(metaFile).Decode(&m); err != nil {
		return Metadata{}, fmt.Errorf("invalid meta file format: %w", err)
	}

	return m, nil
}

func (r *Reader) SaveMetadata() error {
	metaFile, err := os.Create(r.MetaPath)
	if err != nil {
		return err
	}
	defer metaFile.Close()

	var lastModified time.Time
	var expires time.Time
	headers := make(map[string]string)

	if r.res != nil {
		if lm := r.res.Header.Get("Last-Modified"); lm != "" {
			lastModified, _ = time.Parse(time.RFC1123, lm)
		}
		if exp := r.res.Header.Get("Expires"); exp != "" {
			expires, _ = time.Parse(time.RFC1123, exp)
		}
		if r.contentType == "" {
			r.contentType = r.res.Header.Get("Content-Type")
		}

		for _, header := range cachingHeaders {
			if value := r.res.Header.Get(header); value != "" {
				headers[header] = value
			}
		}
	}

	return json.NewEncoder(metaFile).Encode(Metadata{
		LastModified: lastModified.Unix(),
		CachedAt:     time.Now().Unix(),
		ContentType:  r.contentType,
		Expires:      expires.Unix(),
		Headers:      headers,
	})
}

func (r *Reader) Cleanup() {
	os.Remove(r.FilePath)
	os.Remove(r.MetaPath)
}
