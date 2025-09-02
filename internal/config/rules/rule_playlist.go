package rules

import "iptv-gateway/internal/config/types"

type PlaylistRule struct {
	RemoveDuplicates *RemoveDuplicates `yaml:"remove_duplicates,omitempty"`
}

type RemoveDuplicates []struct {
	Patterns    types.RegexpArr `yaml:"patterns"`
	TrimPattern bool            `yaml:"trim_pattern"`
}
