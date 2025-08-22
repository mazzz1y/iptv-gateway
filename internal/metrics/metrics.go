package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	CacheStatusHit  = "hit"
	CacheStatusMiss = "miss"
)

const (
	RequestTypePlaylist = "playlist"
	RequestTypeEPG      = "epg"
	RequestTypeFile     = "file"
)

const (
	FailureReasonGlobalLimit       = "global_limit"
	FailureReasonSubscriptionLimit = "subscription_limit"
	FailureReasonClientLimit       = "client_limit"
	FailureReasonUpstreamError     = "upstream_error"
)

var (
	Registry = prometheus.NewRegistry()

	ClientStreamsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "iptv_client_streams_active",
			Help: "Currently active client streams (gateway to end clients)",
		},
		[]string{"client_name", "subscription_name", "channel_id"},
	)

	BackendStreamsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "iptv_backend_streams_active",
			Help: "Currently active backend streams (gateway to upstream providers)",
		},
		[]string{"subscription_name", "channel_id"},
	)

	StreamsFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_streams_failures_total",
			Help: "Total number of stream failures",
		},
		[]string{"client_name", "subscription_name", "channel_id", "reason"},
	)

	ListingDownload = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_listing_download_count_total",
			Help: "Total number of listing downloads by client and type",
		},
		[]string{"client_name", "listing_type"},
	)

	ProxyRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_proxy_requests_total",
			Help: "Total proxy requests by client, type and cache status",
		},
		[]string{"client_name", "request_type", "cache_status"},
	)
)

func init() {
	Registry.MustRegister(ClientStreamsActive)
	Registry.MustRegister(BackendStreamsActive)
	Registry.MustRegister(StreamsFailures)
	Registry.MustRegister(ListingDownload)
	Registry.MustRegister(ProxyRequests)
	Registry.MustRegister(collectors.NewGoCollector(
		collectors.WithoutGoCollectorRuntimeMetrics(),
	))
}
