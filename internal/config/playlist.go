package config

import (
	"fmt"
	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/config/proxy"
)

type Playlist struct {
	Name    string             `yaml:"name"`
	Sources common.StringOrArr `yaml:"sources"`
	Proxy   proxy.Proxy        `yaml:"proxy,omitempty"`
}

func (p *Playlist) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("playlist name is required")
	}
	if len(p.Sources) == 0 {
		return fmt.Errorf("playlist sources are required")
	}
	for i, source := range p.Sources {
		if source == "" {
			return fmt.Errorf("playlist source[%d] cannot be empty", i)
		}
	}

	return nil
}
