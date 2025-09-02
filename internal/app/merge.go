package app

import (
	"iptv-gateway/internal/config"
)

func mergeArrays[T any](arrays ...[]T) []T {
	result := make([]T, 0)

	for _, array := range arrays {
		result = append(result, array...)
	}

	return result
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
			result.TemplateVars = make(map[string]any, len(handler.TemplateVars))
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
