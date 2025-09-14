package playlist

import "fmt"

type Rule struct {
	RemoveDuplicates *RemoveDuplicatesRule `yaml:"remove_duplicates,omitempty"`
}

func (p *Rule) Validate() error {
	if p.RemoveDuplicates != nil {
		return p.RemoveDuplicates.Validate()
	}
	return fmt.Errorf("playlist rule: at least one rule type must be specified")
}
