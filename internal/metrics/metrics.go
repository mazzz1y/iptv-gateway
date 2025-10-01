package metrics

import (
	"context"
	"iptv-gateway/internal/ctxutil"

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

	playlistStreamsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "iptv_playlist_streams_active",
			Help: "Currently active playlist streams",
		},
		[]string{"playlist_name"},
	)

	clientStreamsActive = NewAutoCleanGauge(
		"iptv_client_streams_active",
		"Currently active client streams",
		[]string{"client_name", "playlist_name", "channel_name"},
	)

	streamsReusedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_streams_reused_total",
			Help: "Total number of reused streams",
		},
		[]string{"playlist_name", "channel_name"},
	)

	streamsFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_streams_failures_total",
			Help: "Total number of stream failures",
		},
		[]string{"client_name", "playlist_name", "channel_name", "reason"},
	)

	listingDownloadTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_listing_downloads_total",
			Help: "Total number of listing downloads by client and type",
		},
		[]string{"client_name", "request_type"},
	)

	proxyRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "iptv_proxy_requests_total",
			Help: "Total proxy requests by client, type and cache status",
		},
		[]string{"client_name", "request_type", "cache_status"},
	)
)

func IncPlaylistStreamsActive(ctx context.Context) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}
	subscriptionName := ctxutil.ProviderName(ctx)
	playlistStreamsActive.WithLabelValues(subscriptionName).Inc()
}

func DecPlaylistStreamsActive(ctx context.Context) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}
	subscriptionName := ctxutil.ProviderName(ctx)
	playlistStreamsActive.WithLabelValues(subscriptionName).Dec()
}

func SetPlaylistStreamsActive(playlistName string, value float64) {
	playlistStreamsActive.WithLabelValues(playlistName).Set(value)
}

func IncClientStreamsActive(ctx context.Context) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}

	channelID := ctxutil.ChannelName(ctx)
	clientName := ctxutil.ClientName(ctx)
	playlistName := ctxutil.ProviderName(ctx)
	clientStreamsActive.WithLabelValues(clientName, playlistName, channelID).Inc()
}

func DecClientStreamsActive(ctx context.Context) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}

	channelID := ctxutil.ChannelName(ctx)
	clientName := ctxutil.ClientName(ctx)
	playlistName := ctxutil.ProviderName(ctx)
	clientStreamsActive.WithLabelValues(clientName, playlistName, channelID).Dec()
}

func IncStreamsReused(ctx context.Context) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}
	subscriptionName := ctxutil.ProviderName(ctx)
	channelID := ctxutil.ChannelName(ctx)
	streamsReusedTotal.WithLabelValues(subscriptionName, channelID).Inc()
}

func IncStreamsFailures(ctx context.Context, reason string) {
	if ctxutil.ChannelHidden(ctx) {
		return
	}

	channelName := ctxutil.ChannelName(ctx)
	clientName := ctxutil.ClientName(ctx)
	playlistName := ctxutil.ProviderName(ctx)
	streamsFailuresTotal.WithLabelValues(clientName, playlistName, channelName, reason).Inc()
}

func IncListingDownload(ctx context.Context) {
	clientName := ctxutil.ClientName(ctx)
	requestType := ctxutil.RequestType(ctx)
	listingDownloadTotal.WithLabelValues(clientName, requestType).Inc()
}

func IncProxyRequests(ctx context.Context, cacheStatus string) {
	clientName := ctxutil.ClientName(ctx)
	requestType := ctxutil.RequestType(ctx)
	proxyRequestsTotal.WithLabelValues(clientName, requestType, cacheStatus).Inc()
}

func init() {
	Registry.MustRegister(clientStreamsActive)
	Registry.MustRegister(playlistStreamsActive)
	Registry.MustRegister(streamsReusedTotal)
	Registry.MustRegister(streamsFailuresTotal)
	Registry.MustRegister(listingDownloadTotal)
	Registry.MustRegister(proxyRequestsTotal)
	Registry.MustRegister(collectors.NewGoCollector(
		collectors.WithoutGoCollectorRuntimeMetrics(),
	))
}
