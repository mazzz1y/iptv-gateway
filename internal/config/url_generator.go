package config

import (
	"fmt"
	"majmun/internal/config/common"
)

type URLGeneratorConfig struct {
	Secret    string          `yaml:"secret"`
	StreamTTL common.Duration `yaml:"stream_ttl"`
	FileTTL   common.Duration `yaml:"file_ttl"`
}

func (u *URLGeneratorConfig) Validate() error {
	if u.Secret == "" {
		return fmt.Errorf("secret is required")
	}
	return nil
}
