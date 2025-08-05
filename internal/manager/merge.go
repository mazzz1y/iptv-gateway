package manager

import (
	"iptv-gateway/internal/config"
)

func mergeExcludes(filters ...config.Excludes) config.Excludes {
	result := config.Excludes{
		Tags:        make(map[string]config.RegexpArr),
		Attrs:       make(map[string]config.RegexpArr),
		ChannelName: make(config.RegexpArr, 0),
	}

	for _, filter := range filters {
		mergeTagFilters(result.Tags, filter.Tags)
		mergeAttrFilters(result.Attrs, filter.Attrs)
		
		if len(filter.ChannelName) > 0 {
			result.ChannelName = append(result.ChannelName, filter.ChannelName...)
		}
	}

	return result
}

func mergeTagFilters(result, source map[string]config.RegexpArr) {
	for key, value := range source {
		if existingArr, exists := result[key]; exists {
			result[key] = append(existingArr, value...)
		} else {
			regexps := make(config.RegexpArr, len(value))
			copy(regexps, value)
			result[key] = regexps
		}
	}
}

func mergeAttrFilters(result, source map[string]config.RegexpArr) {
	for key, value := range source {
		if existingArr, exists := result[key]; exists {
			result[key] = append(existingArr, value...)
		} else {
			regexps := make(config.RegexpArr, len(value))
			copy(regexps, value)
			result[key] = regexps
		}
	}
}

func mergeProxies(proxies ...config.Proxy) config.Proxy {
	result := config.Proxy{}

	for _, proxy := range proxies {
		if proxy.Enabled != nil {
			result.Enabled = proxy.Enabled
		}

		if proxy.ConcurrentStreams > 0 {
			result.ConcurrentStreams = proxy.ConcurrentStreams
		}

		result.Stream = mergeHandlers(result.Stream, proxy.Stream)
		result.Error.Handler = mergeHandlers(result.Error.Handler, proxy.Error.Handler)
		
				result.Error.RateLimitExceeded = mergeHandlers(
			result.Error.Handler, 
			result.Error.RateLimitExceeded, 
			proxy.Error.RateLimitExceeded,
		)
		
		result.Error.LinkExpired = mergeHandlers(
			result.Error.Handler, 
			result.Error.LinkExpired, 
			proxy.Error.LinkExpired,
		)
		
		result.Error.UpstreamError = mergeHandlers(
			result.Error.Handler, 
			result.Error.UpstreamError, 
			proxy.Error.UpstreamError,
		)
	}

	return result
}

func mergeHandlers(handlers ...config.Handler) config.Handler {
	result := config.Handler{}

	for _, handler := range handlers {
		if len(handler.Command) > 0 {
			result.Command = handler.Command
		}

		mergeTemplateVars(&result, handler)
		mergeEnvVars(&result, handler)
	}

	return result
}

func mergeTemplateVars(result *config.Handler, handler config.Handler) {
	if len(handler.TemplateVars) > 0 {
		if result.TemplateVars == nil {
			result.TemplateVars = make(map[string]string)
		}
		for k, v := range handler.TemplateVars {
			result.TemplateVars[k] = v
		}
	}
}

func mergeEnvVars(result *config.Handler, handler config.Handler) {
	if len(handler.EnvVars) > 0 {
		if result.EnvVars == nil {
			result.EnvVars = make(map[string]string)
		}
		for k, v := range handler.EnvVars {
			result.EnvVars[k] = v
		}
	}
}
