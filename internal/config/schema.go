package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr    string         `yaml:"listen_addr"`
	PublicURL     PublicURL      `yaml:"public_url"`
	LogLevel      string         `yaml:"log_level"`
	Secret        string         `yaml:"secret"`
	Cache         CacheConfig    `yaml:"cache"`
	Proxy         Proxy          `yaml:"proxy"`
	Clients       []Client       `yaml:"clients"`
	Subscriptions []Subscription `yaml:"subscriptions"`
	Excludes      Excludes       `yaml:"excludes,omitempty"`
	Presets       []Preset       `yaml:"presets,omitempty"`
}

type CacheConfig struct {
	Path string `yaml:"path"`
	TTL  TTL    `yaml:"ttl"`
}

type TTL time.Duration

type Client struct {
	Name          string      `yaml:"name"`
	Secret        string      `yaml:"secret"`
	Subscriptions StringOrArr `yaml:"subscriptions"`
	Preset        StringOrArr `yaml:"presets,omitempty"`
	Proxy         Proxy       `yaml:"proxy,omitempty"`
	Excludes      Excludes    `yaml:"excludes,omitempty"`
}

type Subscription struct {
	Name     string      `yaml:"name"`
	Playlist StringOrArr `yaml:"playlist"`
	EPG      StringOrArr `yaml:"epg"`
	Proxy    Proxy       `yaml:"proxy"`
	Excludes Excludes    `yaml:"excludes,omitempty"`
}

func (s Subscription) GetName() string {
	return s.Name
}

type Excludes struct {
	Tags        map[string]RegexpArr `yaml:"tags,omitempty"`
	Attrs       map[string]RegexpArr `yaml:"attrs,omitempty"`
	ChannelName RegexpArr            `yaml:"channel_name,omitempty"`
}

type Proxy struct {
	Enabled           *bool   `yaml:"enabled"`
	ConcurrentStreams int64   `yaml:"concurrent_streams"`
	Stream            Handler `yaml:"stream,omitempty"`
	Error             Error   `yaml:"error,omitempty"`
}

func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	var enabled bool
	if err := value.Decode(&enabled); err == nil {
		p.Enabled = &enabled
		return nil
	}

	type proxyYAML Proxy
	return value.Decode((*proxyYAML)(p))
}

type Error struct {
	Handler           `yaml:",inline"`
	UpstreamError     Handler `yaml:"upstream_error"`
	RateLimitExceeded Handler `yaml:"rate_limit_exceeded"`
	LinkExpired       Handler `yaml:"link_expired"`
}

type Handler struct {
	Command      StringOrArr       `yaml:"command,omitempty"`
	TemplateVars map[string]any    `yaml:"template_vars,omitempty"`
	EnvVars      map[string]string `yaml:"env_vars,omitempty"`
}

type Preset struct {
	Name     string   `yaml:"name"`
	Proxy    Proxy    `yaml:"proxy,omitempty"`
	Excludes Excludes `yaml:"excludes,omitempty"`
}

func (p Preset) GetName() string {
	return p.Name
}

type RegexpArr []regexp.Regexp

func (r *RegexpArr) UnmarshalYAML(value *yaml.Node) error {
	patterns, err := unmarshalStringOrArray(value)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
	}

	regexps := make([]regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
		}
		regexps = append(regexps, *compiled)
	}

	*r = regexps
	return nil
}

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

func (t *TTL) UnmarshalYAML(value *yaml.Node) error {
	var ttlStr string
	if err := value.Decode(&ttlStr); err != nil {
		return err
	}

	re := regexp.MustCompile(`^(\d+)([smhdwMy])$`)
	matches := re.FindStringSubmatch(ttlStr)

	if matches == nil {
		return fmt.Errorf("invalid TTL format: %s", ttlStr)
	}

	val, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("invalid TTL value: %s", matches[1])
	}

	unit := matches[2]

	switch unit {
	case "s":
		*t = TTL(time.Duration(val) * time.Second)
	case "m":
		*t = TTL(time.Duration(val) * time.Minute)
	case "h":
		*t = TTL(time.Duration(val) * time.Hour)
	case "d":
		*t = TTL(time.Duration(val) * 24 * time.Hour)
	case "w":
		*t = TTL(time.Duration(val) * 7 * 24 * time.Hour)
	case "M":
		*t = TTL(time.Duration(val) * 30 * 24 * time.Hour)
	case "y":
		*t = TTL(time.Duration(val) * 365 * 24 * time.Hour)
	default:
		return fmt.Errorf("unknown time unit: %s", unit)
	}

	return nil
}

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
