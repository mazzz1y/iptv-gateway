package cache

import "net/http"

type cachingTransport struct {
	cache *Cache
}

func (t *cachingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return t.cache.httpClient.Transport.RoundTrip(req)
	}

	reader, err := t.cache.NewReader(req.Context(), req.URL.String())
	if err != nil {
		return t.cache.httpClient.Transport.RoundTrip(req)
	}

	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Header:        make(http.Header),
		Body:          reader,
		ContentLength: -1,
		Request:       req,
		Close:         false,
	}

	if reader.res != nil {
		cachingHeaders := []string{
			"Cache-Control", "Expires", "Last-Modified",
			"ETag", "Age", "Content-Type",
		}
		for _, header := range cachingHeaders {
			if value := reader.res.Header.Get(header); value != "" {
				resp.Header.Set(header, value)
			}
		}
	} else if cachedHeaders := reader.GetCachedHeaders(); cachedHeaders != nil && len(cachedHeaders) > 0 {
		for key, value := range cachedHeaders {
			resp.Header.Set(key, value)
		}
	}

	return resp, nil
}
