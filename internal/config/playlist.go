package config

import (
	"fmt"
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type Playlist struct {
	Name    string            `yaml:"name"`
	Sources types.StringOrArr `yaml:"sources"`
	Proxy   proxy.Proxy       `yaml:"proxy,omitempty"`
	Rules   []*rules.Rule     `yaml:"rules,omitempty"`
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

	for i, rule := range p.Rules {
		if rule.Type == rules.StoreRule {
			return fmt.Errorf("playlist rules[%d] cannot be a playlist rule", i)
		}
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("playlist rules[%d]: %w", i, err)
		}
	}

	return nil
}
