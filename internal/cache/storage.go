package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var forwardedHeaders = []string{
	"Cache-Control", "Expires", "Last-Modified", "ETag", "Content-Type",
}

func readMetadata(metaPath string) (Metadata, error) {
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

	headers := make(map[string]string, len(forwardedHeaders))

	if r.originResponse != nil {
		for _, header := range forwardedHeaders {
			if value := r.originResponse.Header.Get(header); value != "" {
				headers[header] = value
			}
		}
	}

	return json.NewEncoder(metaFile).Encode(Metadata{
		CachedAt: time.Now().Unix(),
		Headers:  headers,
	})
}

func (r *Reader) Cleanup() {
	os.Remove(r.FilePath)
	os.Remove(r.MetaPath)
}
