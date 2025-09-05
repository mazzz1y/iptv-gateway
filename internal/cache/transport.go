package cache

import (
	"net/http"
)

type cachingTransport struct {
	cache *Cache
}

func (t *cachingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reader, err := t.cache.NewReader(req.Context(), req.URL.String())
	if err != nil {
		return nil, err
	}

	status := http.StatusOK
	resp := &http.Response{
		Status:        http.StatusText(status),
		StatusCode:    status,
		Header:        make(http.Header, len(forwardedHeaders)),
		Body:          reader,
		ContentLength: -1,
		Request:       req,
		Close:         false,
	}

	if reader.originResponse != nil {
		for _, header := range forwardedHeaders {
			if value := reader.originResponse.Header.Get(header); value != "" {
				resp.Header.Set(header, value)
			}
		}
	} else if cachedHeaders := reader.getCachedHeaders(); len(cachedHeaders) > 0 {
		for key, value := range cachedHeaders {
			resp.Header.Set(key, value)
		}
	}

	return resp, nil
}
