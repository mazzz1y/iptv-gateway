package config

import (
	"fmt"
	"majmun/internal/config/common"
)

type CacheConfig struct {
	Path        string             `yaml:"path"`
	TTL         common.Duration    `yaml:"ttl"`
	Retention   common.Duration    `yaml:"retention"`
	Compression bool               `yaml:"compression"`
	HttpHeaders []common.NameValue `yaml:"http_headers"`
}

func (c *CacheConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("cache: path is required")
	}
	if c.TTL <= 0 {
		return fmt.Errorf("cache: TTL must be positive")
	}
	if c.Retention <= 0 {
		return fmt.Errorf("cache: retention must be positive")
	}
	for i, header := range c.HttpHeaders {
		if err := header.Validate(); err != nil {
			return fmt.Errorf("cache: header[%d]: %w", i, err)
		}
	}
	return nil
}
