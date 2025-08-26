package cache

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
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
	statusRenewed
	statusNotFound

	fileExtension = ".gz"
	metaExtension = ".meta"
)

type Metadata struct {
	CachedAt int64             `json:"cached_at"`
	Headers  map[string]string `json:"headers"`
}

type Reader struct {
	URL            string
	Name           string
	FilePath       string
	MetaPath       string
	ReadCloser     io.ReadCloser
	file           *os.File
	gzipWriter     *gzip.Writer
	originResponse *http.Response
	client         *http.Client
	bytesDownload  int64
	expectedSize   int64
	contentType    string
	ttl            time.Duration
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.ReadCloser == nil {
		return 0, io.EOF
	}
	n, err = r.ReadCloser.Read(p)

	if (err == io.EOF || r.bytesDownload == r.expectedSize) && r.originResponse != nil {
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

	if r.originResponse != nil {
		if e := r.originResponse.Body.Close(); e != nil {
			err = e
		}
	}

	if r.ReadCloser != nil {
		if e := r.ReadCloser.Close(); e != nil && err == nil {
			err = e
		}
	}

	if r.gzipWriter != nil {
		if e := r.gzipWriter.Close(); e != nil && err == nil {
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

func (r *Reader) getCachedHeaders() map[string]string {
	meta, err := readMetadata(r.MetaPath)
	if err != nil {
		return nil
	}
	return meta.Headers
}

func (r *Reader) createCacheFile() error {
	os.Remove(r.FilePath)
	file, err := os.Create(r.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}

	r.file = file
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

func (r *Reader) checkCacheStatus() status {
	if _, err := os.Stat(r.MetaPath); err != nil {
		return statusNotFound
	}

	if _, err := os.Stat(r.FilePath); err != nil {
		return statusNotFound
	}

	meta, err := readMetadata(r.MetaPath)
	if err != nil {
		return statusNotFound
	}

	if r.ttl > 0 {
		cachedAt := time.Unix(meta.CachedAt, 0)
		if time.Since(cachedAt) < r.ttl {
			return statusValid
		}
	}

	if exp, ok := meta.Headers["Expires"]; ok {
		if expires, err := time.Parse(time.RFC1123, exp); err == nil && expires.Before(time.Now()) {
			return r.tryRenewal(&meta)
		}
		if err := r.SaveMetadata(); err != nil {
			return statusExpired
		}
		return statusRenewed
	}

	return r.tryRenewal(&meta)
}

func (r *Reader) tryRenewal(meta *Metadata) status {
	var lastModified time.Time
	var etag string

	if lm, ok := meta.Headers["Last-Modified"]; ok {
		if parsedTime, err := time.Parse(time.RFC1123, lm); err == nil {
			lastModified = parsedTime
		}
	}

	if tag, ok := meta.Headers["Etag"]; ok {
		etag = tag
	}

	if !r.isModifiedSince(lastModified, etag) {
		if err := r.SaveMetadata(); err != nil {
			return statusExpired
		}
		return statusRenewed
	}

	return statusExpired
}

func (r *Reader) isModifiedSince(lastModified time.Time, etag string) bool {
	if lastModified.IsZero() && etag == "" {
		return true
	}

	req, err := http.NewRequest("HEAD", r.URL, nil)
	if err != nil {
		return true
	}

	if !lastModified.IsZero() {
		req.Header.Set("If-Modified-Since", lastModified.Format(time.RFC1123))
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	r.originResponse = resp

	if resp.StatusCode == http.StatusNotModified {
		return false
	}

	if resp.StatusCode == http.StatusOK && !lastModified.IsZero() {
		serverLastModified := resp.Header.Get("Last-Modified")
		if serverLastModified != "" {
			if serverTime, err := time.Parse(time.RFC1123, serverLastModified); err == nil {
				return serverTime.After(lastModified)
			}
		}
	}

	return true
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

	r.originResponse = resp
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

	r.originResponse = resp
	r.contentType = resp.Header.Get("Content-Type")

	if r.originResponse.ContentLength > 0 {
		r.expectedSize = r.originResponse.ContentLength
	}

	err = r.createCacheFile()
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
		gzipW, err := gzip.NewWriterLevel(r.file, gzip.BestSpeed)
		if err != nil {
			_ = r.file.Close()
			_ = sc.Close()
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		r.gzipWriter = gzipW
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
	case statusRenewed:
		return "renewed"
	default:
		return "unknown"
	}
}
