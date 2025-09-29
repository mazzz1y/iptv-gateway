package app

import (
	"iptv-gateway/internal/config/common"
	"iptv-gateway/internal/config/proxy"
)

func uniqueNames(names []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, n := range names {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			result = append(result, n)
		}
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
		result.Stream = mergeHandlers(
			result.Stream, p.Stream)

		result.Error.Handler = mergeHandlers(
			result.Error.Handler, p.Error.Handler)

		result.Error.RateLimitExceeded = mergeHandlers(
			result.Error.Handler, result.Error.RateLimitExceeded, p.Error.RateLimitExceeded)

		result.Error.LinkExpired = mergeHandlers(
			result.Error.Handler, result.Error.LinkExpired, p.Error.LinkExpired)

		result.Error.UpstreamError = mergeHandlers(
			result.Error.Handler, result.Error.UpstreamError, p.Error.UpstreamError)
	}

	return result
}

func mergeHandlers(handlers ...proxy.Handler) proxy.Handler {
	result := proxy.Handler{}
	for _, h := range handlers {
		if len(h.Command) > 0 {
			result.Command = h.Command
		}
		mergePairs(&result.TemplateVars, h.TemplateVars)
		mergePairs(&result.EnvVars, h.EnvVars)
	}

	return result
}

func mergePairs[T ~[]common.NameValue](result *T, handler T) {
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
	merged := make([]common.NameValue, 0, len(varMap))
	for name, value := range varMap {
		merged = append(merged, common.NameValue{Name: name, Value: value})
	}
	*result = merged
}
