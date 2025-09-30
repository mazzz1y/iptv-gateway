package channel

import (
	"fmt"
	"iptv-gateway/internal/config/common"
)

type RemoveChannelRule struct {
	Condition *common.Condition `yaml:"condition,omitempty"`
}

func (r *RemoveChannelRule) Validate() error {
	if r.Condition == nil {
		return fmt.Errorf("remove_channel: condition is required")
	}
	if err := r.Condition.Validate(); err != nil {
		return fmt.Errorf("remove_channel: %w", err)
	}
	return nil
}
