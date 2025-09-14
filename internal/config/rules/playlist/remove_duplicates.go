package playlist

import (
	"fmt"
	"iptv-gateway/internal/config/rules"
	"iptv-gateway/internal/config/types"
)

type RemoveDuplicatesRule struct {
	NamePatterns types.RegexpArr     `yaml:"name_patterns,omitempty"`
	AttrPatterns *types.NamePatterns `yaml:"attr,omitempty"`
	TagPatterns  *types.NamePatterns `yaml:"tag,omitempty"`
	When         *rules.Condition    `yaml:"when,omitempty"`
}

func (r *RemoveDuplicatesRule) Validate() error {
	fieldCount := r.countPatternFields()
	if fieldCount != 1 {
		return fmt.Errorf("remove_duplicates: exactly one of name, attr, or tag patterns is required")
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
