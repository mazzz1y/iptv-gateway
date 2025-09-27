package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type MergeChannelsRule struct {
	NamePatterns types.RegexpArr     `yaml:"name_patterns,omitempty"`
	AttrPatterns *types.NamePatterns `yaml:"attr,omitempty"`
	TagPatterns  *types.NamePatterns `yaml:"tag,omitempty"`
	SetField     *types.Template     `yaml:"set_field,omitempty"`
	When         *types.Condition    `yaml:"when,omitempty"`
}

func (r *MergeChannelsRule) Validate() error {
	fieldCount := r.countPatternFields()
	if fieldCount != 1 {
		return fmt.Errorf("merge_channels: exactly one of name or attr patterns is required")
	}

	if err := r.validateWhen(); err != nil {
		return err
	}

	switch {
	case len(r.NamePatterns) > 0:
		return r.validatePatternCount("name", len(r.NamePatterns))
	case r.AttrPatterns != nil:
		if r.AttrPatterns.Name != "tvg-id" {
			return fmt.Errorf("merge_channels: attr patterns only allowed for tvg-id field")
		}
		return r.validateNamedPatterns("attr", r.AttrPatterns)
	case r.TagPatterns != nil:
		return fmt.Errorf("merge_channels: tag patterns are not allowed for merging")
	}

	return nil
}

func (r *MergeChannelsRule) String() string {
	return "merge_channels"
}

func (r *MergeChannelsRule) countPatternFields() int {
	count := 0
	if len(r.NamePatterns) > 0 {
		count++
	}
	if r.AttrPatterns != nil {
		count++
	}
	if r.TagPatterns != nil {
		count++
	}
	return count
}

func (r *MergeChannelsRule) validateNamedPatterns(fieldType string, patterns *types.NamePatterns) error {
	if patterns.Name == "" {
		return fmt.Errorf("merge_channels: %s requires name field", fieldType)
	}
	return r.validatePatternCount(fieldType, len(patterns.Patterns))
}

func (r *MergeChannelsRule) validatePatternCount(fieldType string, count int) error {
	if count < 1 {
		return fmt.Errorf("merge_channels: %s requires at least 1 pattern for merging", fieldType)
	}
	return nil
}

func (r *MergeChannelsRule) validateWhen() error {
	if r.When == nil {
		return nil
	}

	if len(r.When.NamePatterns) > 0 || r.When.Attr != nil || r.When.Tag != nil ||
		len(r.When.Playlists) > 0 || len(r.When.And) > 0 || len(r.When.Or) > 0 {
		return fmt.Errorf("merge_channels: only clients field is allowed in when condition")
	}

	return nil
}
