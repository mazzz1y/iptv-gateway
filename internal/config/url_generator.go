package config

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type URLGeneratorConfig struct {
	Secret    string         `yaml:"secret"`
	StreamTTL types.Duration `yaml:"stream_ttl"`
	FileTTL   types.Duration `yaml:"file_ttl"`
}

func (u *URLGeneratorConfig) Validate() error {
	if u.Secret == "" {
		return fmt.Errorf("secret is required")
	}
	return nil
}
