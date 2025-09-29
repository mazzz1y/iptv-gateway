package common

type Condition struct {
	Selector  *Selector   `yaml:"selector,omitempty"`
	Patterns  RegexpArr   `yaml:"patterns,omitempty"`
	Clients   StringOrArr `yaml:"clients,omitempty"`
	Playlists StringOrArr `yaml:"playlists,omitempty"`
	And       []Condition `yaml:"and,omitempty"`
	Or        []Condition `yaml:"or,omitempty"`
	Invert    bool        `yaml:"invert,omitempty"`
}

func (c *Condition) Validate() error {
	if c.Selector != nil {
		if err := c.Selector.Validate(); err != nil {
			return err
		}
	}

	for _, and := range c.And {
		if err := and.Validate(); err != nil {
			return err
		}
	}

	for _, or := range c.Or {
		if err := or.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Condition) IsEmpty() bool {
	return c.Selector == nil && len(c.Patterns) == 0 && len(c.Clients) == 0 &&
		len(c.Playlists) == 0 && len(c.And) == 0 && len(c.Or) == 0 && !c.Invert
}
