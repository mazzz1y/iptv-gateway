package types

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

type NameTemplate struct {
	Name     string    `yaml:"name"`
	Template *Template `yaml:"template"`
}

func (nt *NameTemplate) Validate() error {
	if nt.Name == "" {
		return fmt.Errorf("name is required")
	}
	if nt.Template == nil {
		return fmt.Errorf("template is required")
	}
	return nil
}

type NamePatterns struct {
	Name     string    `yaml:"name"`
	Patterns RegexpArr `yaml:"patterns"`
}

func (np *NamePatterns) Validate() error {
	if np.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(np.Patterns) == 0 {
		return fmt.Errorf("patterns are required")
	}
	return nil
}
