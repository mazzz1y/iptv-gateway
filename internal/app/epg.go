package app

import (
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/urlgen"
)

type EPG struct {
	name string

	sources []string

	urlGenerator *urlgen.Generator
	proxyConfig  config.Proxy
}

func NewEPG(
	name string, urlGen urlgen.Generator,
	sources []string,
	proxy config.Proxy) (*EPG, error) {

	return &EPG{
		name:         name,
		urlGenerator: &urlGen,
		sources:      sources,
		proxyConfig:  proxy,
	}, nil
}

func (es *EPG) Name() string {
	return es.name
}

func (es *EPG) Type() string {
	return "epg"
}

func (es *EPG) EPGs() []string {
	return es.sources
}

func (es *EPG) URLGenerator() *urlgen.Generator {
	return es.urlGenerator
}

func (es *EPG) IsProxied() bool {
	return es.proxyConfig.Enabled != nil && *es.proxyConfig.Enabled
}

func (es *EPG) ProxyConfig() config.Proxy {
	return es.proxyConfig
}
