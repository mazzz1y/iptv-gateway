package config

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type ServerConfig struct {
	ListenAddr  string    `yaml:"listen_addr"`
	MetricsAddr string    `yaml:"metrics_addr"`
	PublicURL   types.URL `yaml:"public_url"`
}

func (s *ServerConfig) Validate() error {
	if s.ListenAddr == "" {
		return fmt.Errorf("listen_addr is required")
	}
	if s.PublicURL.String() == "" {
		return fmt.Errorf("public_url is required")
	}
	return nil
}
