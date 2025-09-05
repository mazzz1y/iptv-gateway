package config

import "iptv-gateway/internal/config/types"

type CacheConfig struct {
	Path        string         `yaml:"path"`
	TTL         types.Duration `yaml:"ttl"`
	Retention   types.Duration `yaml:"retention"`
	Compression bool           `yaml:"compression"`
}
