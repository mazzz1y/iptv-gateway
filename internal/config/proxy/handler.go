package proxy

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type Handler struct {
	Command      types.StringOrArr `yaml:"command,omitempty"`
	TemplateVars []types.NameValue `yaml:"template_vars,omitempty"`
	EnvVars      []types.NameValue `yaml:"env_vars,omitempty"`
}

func (h *Handler) Validate() error {
	for i, templateVar := range h.TemplateVars {
		if err := templateVar.Validate(); err != nil {
			return fmt.Errorf("template_vars[%d]: %w", i, err)
		}
	}

	for i, envVar := range h.EnvVars {
		if err := envVar.Validate(); err != nil {
			return fmt.Errorf("env_vars[%d]: %w", i, err)
		}
	}

	return nil
}
