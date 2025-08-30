package types

import (
	"fmt"
	"net/url"

	"gopkg.in/yaml.v3"
)

type PublicURL url.URL

func (pu *PublicURL) UnmarshalYAML(value *yaml.Node) error {
	var urlStr string
	if err := value.Decode(&urlStr); err != nil {
		return err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return fmt.Errorf("public_url must contain a host")
	}

	*pu = PublicURL(*u)
	return nil
}

func (pu *PublicURL) String() string {
	u := url.URL(*pu)
	return u.String()
}
