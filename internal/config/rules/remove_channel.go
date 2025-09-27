package rules

import (
	"errors"
	"iptv-gateway/internal/config/types"
)

type RemoveChannelRule struct {
	When *types.Condition `yaml:"when,omitempty"`
}

func (r *RemoveChannelRule) Validate() error {
	if r.When == nil {
		return errors.New("remove_channel: when is required")
	}
	return r.When.Validate()
}
