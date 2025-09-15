package channel

import (
	"errors"
	"iptv-gateway/internal/config/rules"
)

type RemoveChannelRule struct {
	When *rules.Condition `yaml:"when,omitempty"`
}

func (r *RemoveChannelRule) Validate() error {
	if r.When == nil {
		return errors.New("when is required")
	}
	return nil
}
