package listing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func GenerateHashID(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p)
	}
	hash := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("%x", hash[:4])
}

func CreateReader(ctx context.Context, httpClient HTTPClient, resourceURL string) (io.ReadCloser, error) {
	if isURL(resourceURL) {
		req, err := http.NewRequestWithContext(ctx, "GET", resourceURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	reader, err := openLocalFile(resourceURL)
	if err != nil {
		return nil, err
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
