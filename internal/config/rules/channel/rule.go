package channel

import "fmt"

type Rule struct {
	SetField      *SetFieldRule      `yaml:"set_field,omitempty"`
	RemoveField   *RemoveFieldRule   `yaml:"remove_field,omitempty"`
	RemoveChannel *RemoveChannelRule `yaml:"remove_channel,omitempty"`
	MarkHidden    *MarkHiddenRule    `yaml:"mark_hidden,omitempty"`
}

func (c *Rule) Validate() error {
	ruleCount := 0
	if c.SetField != nil {
		ruleCount++
		if err := c.SetField.Validate(); err != nil {
			return err
		}
	}
	if c.RemoveField != nil {
		ruleCount++
		if err := c.RemoveField.Validate(); err != nil {
			return err
		}
	}
	if c.RemoveChannel != nil {
		ruleCount++
	}
	if c.MarkHidden != nil {
		ruleCount++
	}

	if ruleCount != 1 {
		return fmt.Errorf("channel rule: exactly one rule type must be specified")
	}
	return nil
}
