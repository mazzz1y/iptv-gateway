package listing

import (
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type Decoder interface {
	Decode() (any, error)
}
