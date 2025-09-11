package config

import (
	"iptv-gateway/internal/config/types"

	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Enabled           *bool   `yaml:"enabled"`
	ConcurrentStreams int64   `yaml:"concurrency"`
	Stream            Handler `yaml:"stream,omitempty"`
	Error             Error   `yaml:"error,omitempty"`
}

func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	var enabled bool
	if err := value.Decode(&enabled); err == nil {
		p.Enabled = &enabled
		return nil
	}

	type proxyYAML Proxy
	return value.Decode((*proxyYAML)(p))
}

type Error struct {
	Handler           `yaml:",inline"`
	UpstreamError     Handler `yaml:"upstream_error"`
	RateLimitExceeded Handler `yaml:"rate_limit_exceeded"`
	LinkExpired       Handler `yaml:"link_expired"`
}

type Handler struct {
	Command      types.StringOrArr `yaml:"command,omitempty"`
	TemplateVars []types.NameValue `yaml:"template_vars,omitempty"`
	EnvVars      []types.NameValue `yaml:"env_vars,omitempty"`
}
