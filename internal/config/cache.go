package config

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type CacheConfig struct {
	Path        string         `yaml:"path"`
	TTL         types.Duration `yaml:"ttl"`
	Retention   types.Duration `yaml:"retention"`
	Compression bool           `yaml:"compression"`
}

func (c *CacheConfig) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("cache path is required")
	}
	if c.TTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}
	if c.Retention <= 0 {
		return fmt.Errorf("cache retention must be positive")
	}
	return nil
}
