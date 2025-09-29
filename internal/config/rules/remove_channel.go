package rules

import (
	"errors"
	"iptv-gateway/internal/config/common"
)

type RemoveChannelRule struct {
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (r *RemoveChannelRule) Validate() error {
	if r.Condition == nil {
		return errors.New("remove_channel: condition is required")
	}
	return r.Condition.Validate()
}
