package proxy

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type Handler struct {
	Command      common.StringOrArr `yaml:"command,omitempty"`
	TemplateVars []common.NameValue `yaml:"template_vars,omitempty"`
	EnvVars      []common.NameValue `yaml:"env_vars,omitempty"`
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
