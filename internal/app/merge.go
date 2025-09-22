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

	for _, p := range proxies {
		if p.Enabled != nil {
			result.Enabled = p.Enabled
		}

		if p.ConcurrentStreams > 0 {
			result.ConcurrentStreams = p.ConcurrentStreams
		}

		result.Stream = mergeHandlers(result.Stream, p.Stream)
		result.Error.Handler = mergeHandlers(result.Error.Handler, p.Error.Handler)

		result.Error.RateLimitExceeded = mergeHandlers(
			result.Error.Handler,
			result.Error.RateLimitExceeded,
			p.Error.RateLimitExceeded,
		)

		result.Error.LinkExpired = mergeHandlers(
			result.Error.Handler,
			result.Error.LinkExpired,
			p.Error.LinkExpired,
		)

		result.Error.UpstreamError = mergeHandlers(
			result.Error.Handler,
			result.Error.UpstreamError,
			p.Error.UpstreamError,
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

func uniqueNames(names []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, name := range names {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	return result
}
