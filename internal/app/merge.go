package app

import (
	"iptv-gateway/internal/config/proxy"
	"iptv-gateway/internal/config/types"
)

func mergeArrays[T any](arrays ...[]T) []T {
	result := make([]T, 0)

	for _, array := range arrays {
		result = append(result, array...)
	}

	return result
}

func mergeProxies(proxies ...proxy.Proxy) proxy.Proxy {
	result := proxy.Proxy{}

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

func mergeHandlers(handlers ...proxy.Handler) proxy.Handler {
	result := proxy.Handler{}

	for _, handler := range handlers {
		if len(handler.Command) > 0 {
			result.Command = handler.Command
		}
		mergePairs(&result.TemplateVars, handler.TemplateVars)
		mergePairs(&result.EnvVars, handler.EnvVars)
	}

	return result
}

func mergePairs[T ~[]types.NameValue](result *T, handler T) {
	if len(handler) == 0 {
		return
	}

	varMap := make(map[string]string, len(*result)+len(handler))

	for _, v := range *result {
		varMap[v.Name] = v.Value
	}
	for _, v := range handler {
		varMap[v.Name] = v.Value
	}

	merged := make([]types.NameValue, 0, len(varMap))
	for name, value := range varMap {
		merged = append(merged, types.NameValue{Name: name, Value: value})
	}

	*result = merged
}
