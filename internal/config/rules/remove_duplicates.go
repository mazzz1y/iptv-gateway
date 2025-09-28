package rules

import (
	"fmt"
	"iptv-gateway/internal/config/types"
)

type RemoveDuplicatesRule struct {
	NamePatterns types.RegexpArr         `yaml:"name_patterns,omitempty"`
	AttrPatterns *types.NamePatterns     `yaml:"attr,omitempty"`
	TagPatterns  *types.NamePatterns     `yaml:"tag,omitempty"`
	SetField     *types.SetFieldTemplate `yaml:"set_field,omitempty"`
	When         *types.Condition        `yaml:"when,omitempty"`
}

func (r *RemoveDuplicatesRule) Validate() error {
	fieldCount := r.countPatternFields()
	if fieldCount != 1 {
		return fmt.Errorf("remove_duplicates: exactly one of name, attr, or tag patterns is required")
	}

	if err := r.validateWhen(); err != nil {
		return err
	}

	if r.SetField != nil {
		if err := r.SetField.Validate(); err != nil {
			return fmt.Errorf("remove_duplicates: %w", err)
		}
	}

	switch {
	case len(r.NamePatterns) > 0:
		return r.validatePatternCount("name", len(r.NamePatterns))
	case r.AttrPatterns != nil:
		return r.validateNamedPatterns("attr", r.AttrPatterns)
	case r.TagPatterns != nil:
		return r.validateNamedPatterns("tag", r.TagPatterns)
	}

	return nil
}

func (r *RemoveDuplicatesRule) String() string {
	return "remove_duplicates"
}

func (r *RemoveDuplicatesRule) countPatternFields() int {
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

func (r *RemoveDuplicatesRule) validateNamedPatterns(fieldType string, patterns *types.NamePatterns) error {
	if patterns.Name == "" {
		return fmt.Errorf("remove_duplicates: %s requires name field", fieldType)
	}
	return r.validatePatternCount(fieldType, len(patterns.Patterns))
}

func (r *RemoveDuplicatesRule) validatePatternCount(fieldType string, count int) error {
	if count < 2 {
		return fmt.Errorf("remove_duplicates: %s requires at least 2 patterns for duplicate detection", fieldType)
	}
	return nil
}

func (r *RemoveDuplicatesRule) validateWhen() error {
	if r.When == nil {
		return nil
	}

	if len(r.When.NamePatterns) > 0 || r.When.Attr != nil || r.When.Tag != nil ||
		len(r.When.Playlists) > 0 || len(r.When.And) > 0 || len(r.When.Or) > 0 {
		return fmt.Errorf("remove_duplicates: only clients field is allowed in when condition")
	}

	return r.When.Validate()
}
