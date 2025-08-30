package types

import "gopkg.in/yaml.v3"

type StringOrArr []string

func (s *StringOrArr) UnmarshalYAML(node *yaml.Node) error {
	val, err := unmarshalStringOrArray(node)
	if err != nil {
		return err
	}

	*s = val
	return nil
}

func unmarshalStringOrArray(node *yaml.Node) ([]string, error) {
	var singleValue string
	var err error
	if err = node.Decode(&singleValue); err == nil {
		return []string{singleValue}, nil
	}

	var multipleValues []string
	if err = node.Decode(&multipleValues); err == nil {
		return multipleValues, nil
	}

	return nil, err
}
