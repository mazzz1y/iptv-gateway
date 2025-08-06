package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
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
}

type CacheConfig struct {
	Path string `yaml:"path"`
	TTL  TTL    `yaml:"ttl"`
}

type TTL time.Duration

type Client struct {
	Name          string              `yaml:"name"`
	Secret        string              `yaml:"secret"`
	Subscriptions ClientSubscriptions `yaml:"subscriptions"`
	Proxy         Proxy               `yaml:"proxy"`
	Excludes      Excludes            `yaml:"excludes,omitempty"`
}

type Subscription struct {
	Name     string         `yaml:"name"`
	Playlist PlaylistSource `yaml:"playlist"`
	EPG      EPGSource      `yaml:"epg"`
	Proxy    Proxy          `yaml:"proxy"`
	Excludes Excludes       `yaml:"excludes,omitempty"`
}

type Excludes struct {
	Tags        map[string]RegexpArr `yaml:"tags,omitempty"`
	Attrs       map[string]RegexpArr `yaml:"attrs,omitempty"`
	ChannelName RegexpArr            `yaml:"channel_name,omitempty"`
}

type Proxy struct {
	Enabled           *bool   `yaml:"enabled"`
	ConcurrentStreams int64   `yaml:"concurrent_streams"`
	Stream            Handler `yaml:"stream"`
	Error             Error   `yaml:"error"`
}

type Error struct {
	Handler           `yaml:",inline"`
	UpstreamError     Handler `yaml:"upstream_error"`
	RateLimitExceeded Handler `yaml:"rate_limit_exceeded"`
	LinkExpired       Handler `yaml:"link_expired"`
}

type Handler struct {
	Command      Command           `yaml:"command,omitempty"`
	TemplateVars map[string]string `yaml:"template_vars,omitempty"`
	EnvVars      map[string]string `yaml:"env_vars,omitempty"`
}

type Command []string

type PlaylistSource []string
type EPGSource []string
type ClientSubscriptions []string
type RegexpArr []regexp.Regexp

type PublicURL url.URL

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

func (p *Command) UnmarshalYAML(value *yaml.Node) error {
	*p = unmarshalStringOrArray(value)
	return nil
}

func (p *PlaylistSource) UnmarshalYAML(value *yaml.Node) error {
	*p = unmarshalStringOrArray(value)
	return nil
}

func (e *EPGSource) UnmarshalYAML(value *yaml.Node) error {
	*e = unmarshalStringOrArray(value)
	return nil
}

func (s *ClientSubscriptions) UnmarshalYAML(value *yaml.Node) error {
	*s = unmarshalStringOrArray(value)
	return nil
}

func (r *RegexpArr) UnmarshalYAML(value *yaml.Node) error {
	patterns := unmarshalStringOrArray(value)
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

func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	var enabled bool
	if err := value.Decode(&enabled); err == nil {
		p.Enabled = &enabled
		return nil
	}

	type proxyYAML Proxy
	return value.Decode((*proxyYAML)(p))
}

func unmarshalStringOrArray(node *yaml.Node) []string {
	var singleValue string
	if err := node.Decode(&singleValue); err == nil {
		return []string{singleValue}
	}

	var multipleValues []string
	if err := node.Decode(&multipleValues); err == nil {
		return multipleValues
	}

	return nil
}
