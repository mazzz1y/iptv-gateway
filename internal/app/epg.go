package app

import (
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/shell"
	"iptv-gateway/internal/urlgen"
)

type EPG struct {
	name         string
	sources      []string
	urlGenerator *urlgen.Generator
	proxyConfig  proxy.Proxy
}

func NewEPGProvider(
	name string, urlGen *urlgen.Generator, sources []string, proxy proxy.Proxy) (*EPG, error) {
	return &EPG{
		name:         name,
		urlGenerator: urlGen,
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

func (es *EPG) ProxyConfig() proxy.Proxy {
	return es.proxyConfig
}

func (es *EPG) ExpiredLinkStreamer() *shell.Streamer {
	return nil
}
