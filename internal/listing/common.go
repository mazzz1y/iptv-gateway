package listing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func GenerateTvgID(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash[:4])
}

func CreateReader(ctx context.Context, httpClient HTTPClient, resourceURL string) (io.ReadCloser, error) {
	if isURL(resourceURL) {
		req, err := http.NewRequestWithContext(ctx, "GET", resourceURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", resourceURL, err)
		}

		return resp.Body, nil
	}

	reader, err := openLocalFile(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file %s: %w", resourceURL, err)
	}

	return reader, nil
}

func isURL(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

func openLocalFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}
