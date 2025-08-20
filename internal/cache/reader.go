package cache

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"iptv-gateway/internal/constant"
	"iptv-gateway/internal/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type status int

const (
	statusValid status = iota
	statusExpired
	statusNotFound

	fileExtension = ".gz"
	metaExtension = ".meta"
)

type Metadata struct {
	LastModified int64             `json:"last_modified"`
	CachedAt     int64             `json:"cached_at"`
	ContentType  string            `json:"content_type"`
	Expires      int64             `json:"expires"`
	Headers      map[string]string `json:"headers"`
}

type Reader struct {
	URL           string
	Name          string
	FilePath      string
	MetaPath      string
	ReadCloser    io.ReadCloser
	file          *os.File
	writer        *gzip.Writer
	res           *http.Response
	client        *http.Client
	ttl           time.Duration
	bytesDownload int64
	expectedSize  int64
	contentType   string
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.ReadCloser == nil {
		return 0, io.EOF
	}
	n, err = r.ReadCloser.Read(p)

	if (err == io.EOF || r.bytesDownload == r.expectedSize) && r.res != nil {
		if r.isDownloadComplete() {
			err := r.SaveMetadata()
			if err != nil {
				return 0, err
			}
		} else {
			r.Cleanup()
		}
	}
	return n, err
}

func (r *Reader) Close() error {
	var err error

	if r.res != nil {
		if e := r.res.Body.Close(); e != nil {
			err = e
		}
	}

	if r.ReadCloser != nil {
		if e := r.ReadCloser.Close(); e != nil && err == nil {
			err = e
		}
	}

	if r.writer != nil {
		if e := r.writer.Close(); e != nil && err == nil {
			err = e
		}
	}

	if r.file != nil {
		if e := r.file.Close(); e != nil && err == nil {
			err = e
		}
	}

	return err
}

func (r *Reader) GetCachedHeaders() map[string]string {
	meta, err := ReadMetadata(r.MetaPath)
	if err != nil {
		return nil
	}
	return meta.Headers
}

func (r *Reader) CreateCacheFile() error {
	os.Remove(r.FilePath)
	file, err := os.Create(r.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}

	r.file = file
	if r.res.ContentLength > 0 {
		r.expectedSize = r.res.ContentLength
	}

	return nil
}

func (r *Reader) isGzippedContent(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return contentType == "application/gzip" ||
		contentType == "application/x-gzip" ||
		resp.Header.Get("Content-Encoding") == "gzip" ||
		strings.HasSuffix(r.URL, ".gz")
}

func (r *Reader) isDownloadComplete() bool {
	return r.expectedSize <= 0 || r.bytesDownload == r.expectedSize
}

func (r *Reader) isModifiedSince(lastModified time.Time) bool {
	if lastModified.IsZero() {
		return true
	}

	req, err := http.NewRequest("HEAD", r.URL, nil)
	if err != nil {
		return true
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return true
	}

	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		currentLastModified, err := time.Parse(time.RFC1123, lm)
		if err == nil {
			return currentLastModified.After(lastModified)
		}
	}

	return false
}

func (r *Reader) checkCacheStatus() status {
	_, err := os.Stat(r.MetaPath)
	if err != nil {
		return statusNotFound
	}

	_, err = os.Stat(r.FilePath)
	if err != nil {
		return statusNotFound
	}

	meta, err := ReadMetadata(r.MetaPath)
	if err != nil {
		return statusNotFound
	}

	r.contentType = meta.ContentType
	lastModified := time.Unix(meta.LastModified, 0)
	cachedAt := time.Unix(meta.CachedAt, 0)
	expires := time.Unix(meta.Expires, 0)

	if r.ttl > 0 {
		if time.Since(cachedAt) > r.ttl {
			if !r.isModifiedSince(lastModified) {
				if err := r.SaveMetadata(); err != nil {
					return statusExpired
				}
				return statusValid
			}
			return statusExpired
		}
		return statusValid
	}

	if !expires.IsZero() && expires.Before(time.Now()) {
		return statusExpired
	}

	return statusValid
}

func (r *Reader) newCachedReader() (io.ReadCloser, error) {
	file, err := os.Open(r.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cached file: %w", err)
	}

	gzipR, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	r.file = file
	return ioutil.NewReaderWithCloser(gzipR, gzipR.Close), nil
}

func (r *Reader) newDirectReader(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	r.res = resp
	r.contentType = resp.Header.Get("Content-Type")

	if r.isGzippedContent(resp) {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return ioutil.NewReaderWithCloser(gzipReader, gzipReader.Close), nil
	} else {
		return resp.Body, nil
	}
}

func (r *Reader) newCachingReader(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	r.res = resp
	r.contentType = resp.Header.Get("Content-Type")

	err = r.CreateCacheFile()
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	var reader io.ReadCloser
	if r.isGzippedContent(resp) {
		sc := ioutil.NewCountReadCloser(resp.Body, &r.bytesDownload)
		tee := io.TeeReader(sc, r.file)
		gzipReader, err := gzip.NewReader(tee)
		if err != nil {
			_ = r.file.Close()
			_ = sc.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		reader = ioutil.NewReaderWithCloser(gzipReader, gzipReader.Close)
	} else {
		sc := ioutil.NewCountReadCloser(resp.Body, &r.bytesDownload)
		gzipW, err := gzip.NewWriterLevel(r.file, constant.GzipLevel)
		if err != nil {
			_ = r.file.Close()
			_ = sc.Close()
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		r.writer = gzipW
		reader = ioutil.NewReaderWithCloser(io.TeeReader(sc, gzipW), gzipW.Close)
	}

	return reader, nil
}

func formatCacheStatus(status status) string {
	switch status {
	case statusValid:
		return "cached"
	case statusExpired:
		return "expired"
	case statusNotFound:
		return "not found"
	default:
		return "unknown"
	}
}
