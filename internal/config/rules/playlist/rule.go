package playlist

import "fmt"

type Rule struct {
	RemoveDuplicates *RemoveDuplicatesRule `yaml:"remove_duplicates,omitempty"`
	SortRule         *SortRule             `yaml:"sort,omitempty"`
}

func (p *Rule) Validate() error {
	if p.RemoveDuplicates != nil {
		return p.RemoveDuplicates.Validate()
	}
	if p.SortRule != nil {
		return p.SortRule.Validate()
	}
	return fmt.Errorf("playlist rule: at least one rule type must be specified")
}
