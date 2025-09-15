package proxy

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Enabled           *bool   `yaml:"enabled"`
	ConcurrentStreams int64   `yaml:"concurrency"`
	Stream            Handler `yaml:"stream,omitempty"`
	Error             Error   `yaml:"error,omitempty"`
}

func (p *Proxy) Validate() error {
	if p.ConcurrentStreams < 0 {
		return fmt.Errorf("proxy concurrent streams cannot be negative")
	}

	if err := p.Stream.Validate(); err != nil {
		return fmt.Errorf("proxy stream handler: %w", err)
	}

	if err := p.Error.Validate(); err != nil {
		return fmt.Errorf("proxy error handler: %w", err)
	}

	return nil
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
