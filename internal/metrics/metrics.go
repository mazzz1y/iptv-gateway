package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	CacheStatusHit     = "hit"
	CacheStatusMiss    = "miss"
	CacheStatusRenewed = "renewed"
)

const (
	RequestTypePlaylist = "playlist"
	RequestTypeEPG      = "epg"
	RequestTypeFile     = "file"
)

const (
	FailureReasonGlobalLimit   = "global_limit"
	FailureReasonPlaylistLimit = "playlist_limit"
	FailureReasonClientLimit   = "client_limit"
	FailureReasonUpstreamError = "upstream_error"
)

var (
	Registry = prometheus.NewRegistry()

	PlaylistStreamsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "iptv_playlist_streams_active",
			Help: "Currently active playlist streams",
		},
		[]string{"playlist_name"},
	)

	ClientStreamsActive = NewAutoCleanGauge(
		"iptv_client_streams_active",
		"Currently active client streams",
		[]string{"client_name", "playlist_name", "channel_id"},
	)

	StreamsReusedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_streams_reused_total",
			Help: "Total number of reused streams",
		},
		[]string{"playlist_name", "channel_id"},
	)

	StreamsFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_streams_failures_total",
			Help: "Total number of stream failures",
		},
		[]string{"client_name", "playlist_name", "channel_id", "reason"},
	)

	ListingDownloadTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_listing_downloads_total",
			Help: "Total number of listing downloads by client and type",
		},
		[]string{"client_name", "listing_type"},
	)

	ProxyRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_proxy_requests_total",
			Help: "Total proxy requests by client, type and cache status",
		},
		[]string{"client_name", "request_type", "cache_status"},
	)
)

func init() {
	Registry.MustRegister(ClientStreamsActive)
	Registry.MustRegister(PlaylistStreamsActive)
	Registry.MustRegister(StreamsReusedTotal)
	Registry.MustRegister(StreamsFailuresTotal)
	Registry.MustRegister(ListingDownloadTotal)
	Registry.MustRegister(ProxyRequestsTotal)
	Registry.MustRegister(collectors.NewGoCollector(
		collectors.WithoutGoCollectorRuntimeMetrics(),
	))
}
