package common

import "fmt"

type NameValue struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func (nv *NameValue) Validate() error {
	if nv.Name == "" {
		return fmt.Errorf("name is required")
	}
	if nv.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

type NamePatterns struct {
	Name     string    `yaml:"name"`
	Patterns RegexpArr `yaml:"patterns"`
}
