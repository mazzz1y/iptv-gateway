package config

import "iptv-gateway/internal/config/types"

type EPG struct {
	Name    string            `yaml:"name"`
	Sources types.StringOrArr `yaml:"source"`
	Proxy   Proxy             `yaml:"proxy,omitempty"`
}
