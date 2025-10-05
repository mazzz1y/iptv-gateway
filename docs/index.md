---
hide:
  - navigation
  - toc
---
<div style="max-width: 850px; margin: 0 auto;" markdown>

# IPTV Gateway

<div style="display: flex; align-items: center; gap: 1em; flex-wrap: wrap;">
  <img src="assets/logo-tv.svg" alt="logo" width="100"/>
  <div style="flex: 1; min-width: 250px;">
    <strong>A minimal, functional IPTV gateway for your home TVs.</strong><br/>
    Transform and proxy your M3U playlists, EPG, and video streams through a single entry point.
    Configure playlists exactly how each client needs them, declaratively.
  </div>
</div>

<style>
@media (max-width: 500px) {
  div[style*="flex-wrap"] {
    flex-direction: column;
    text-align: center;
  }
}
</style>
---

![Diagram](./assets/diagram.svg)

### :material-playlist-music: Playlist Control

Take full control of your M3U playlists. Filter out unwanted channels, sort by your preferences, rename groups, or
modify any field. The gateway processes everything on-the-fly.

Exact channel duplicates with the same `tvg-id` or name will be merged into a single channel using fallback logic. The first channel has higher priority, and the others will act as fallbacks if the first stream becomes unavailable for any reason.

EPG data streams directly to clients without loading entire XML files into memory. Attach multiple EPG sources and the
gateway intelligently filters to include only relevant program data.

### :material-shield-lock: Smart Proxying

Generate encrypted URLs with an optional TTL that completely hide your original sources. Your clients get secure links they can't decrypt.

The gateway operates statelessly — you can rename channels, modify playlists, or update configurations without breaking existing client links. Everything just keeps working.

When multiple clients request the same stream, the gateway automatically demultiplexes from a single upstream connection, reducing bandwidth and server load.

Exact channel duplicates with the same `tvg-id` or name are merged into a single channel using fallback logic. The first channel has higher priority, and the others act as fallbacks if the first stream becomes unavailable for any reason.

**Need to transcode?** Configure any FFmpeg command or external tool for stream processing. The gateway handles the
pipeline, you define the transformation.

Mix and match rules to create exactly the experience each client requires.

<div class="grid cards" markdown>

- :material-cached:{ .lg .middle } [Intelligent Caching](config/cache.md)

    ---

    All remote files is cached on disk. The gateway reuses cached data whenever possible, reducing upstream requests.

- :material-valve:{ .lg .middle } [Rate Limiting](config/proxy.md)

    ---

    Control concurrent streams at any level — limit total gateway streams, restrict per subscription, or set individual client limits.

- :material-alert-circle:{ .lg .middle } [Custom Error Handling](config/proxy.md)

    ---

    Define meaningful error responses for different scenarios. Create specific messages for rate limit violations, expired links, or upstream failures.

- :material-chart-line:{ .lg .middle } [Built-in Monitoring](metrics.md)

    ---

    Export Prometheus metrics to monitor usage patterns per client.

</div>