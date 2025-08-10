package constant

import (
	"compress/gzip"
	"time"
)

const (
	GzipLevel        = gzip.BestSpeed
	SemaphoreTimeout = 6 * time.Second
)
