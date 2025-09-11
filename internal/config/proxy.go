package config

import (
	"iptv-gateway/internal/config/types"

	"gopkg.in/yaml.v3"
)

type EnvNameValue struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

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
	TemplateVars []EnvNameValue    `yaml:"template_vars,omitempty"`
	EnvVars      []EnvNameValue    `yaml:"env_vars,omitempty"`
}
