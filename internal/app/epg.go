package app

import (
	"iptv-gateway/internal/config"
	"iptv-gateway/internal/urlgen"
)

type EPGSubscription struct {
	name string

	sources []string

	urlGenerator *urlgen.Generator
	proxyConfig  config.Proxy
}

func NewEPGSubscription(
	name string, urlGen urlgen.Generator,
	sources []string,
	proxy config.Proxy) (*EPGSubscription, error) {

	return &EPGSubscription{
		name:         name,
		urlGenerator: &urlGen,
		sources:      sources,
		proxyConfig:  proxy,
	}, nil
}

func (es *EPGSubscription) Name() string {
	return es.name
}

func (es *EPGSubscription) Type() string {
	return "epg"
}

func (es *EPGSubscription) EPGs() []string {
	return es.sources
}

func (es *EPGSubscription) URLGenerator() *urlgen.Generator {
	return es.urlGenerator
}

func (es *EPGSubscription) IsProxied() bool {
	return es.proxyConfig.Enabled != nil && *es.proxyConfig.Enabled
}

func (es *EPGSubscription) ProxyConfig() config.Proxy {
	return es.proxyConfig
}
