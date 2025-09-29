package config

import (
	"fmt"
	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/config/proxy"
)

type EPG struct {
	Name    string             `yaml:"name"`
	Sources common.StringOrArr `yaml:"sources"`
	Proxy   proxy.Proxy        `yaml:"proxy,omitempty"`
}

func (e *EPG) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("EPG name is required")
	}
	if len(e.Sources) == 0 {
		return fmt.Errorf("EPG sources are required")
	}
	for i, source := range e.Sources {
		if source == "" {
			return fmt.Errorf("EPG source[%d] cannot be empty", i)
		}
	}
	return nil
}
