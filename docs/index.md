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

### :material-playlist-music: Playlist Control

Take full control of your M3U playlists. Filter out unwanted channels, sort by your preferences, rename groups, or
modify any field. The gateway processes everything on-the-fly.

Duplicate channels with different quality markers (4K/HD/SD) are automatically detected and can be consolidated based on
your rules.

EPG data streams directly to clients without loading entire XML files into memory. Attach multiple EPG sources and the
gateway intelligently filters to include only relevant program data.

### :material-shield-lock: Smart Proxying

Generate encrypted URLs with optional TTL that completely hide your original sources. Your clients receive secure links
they can't decrypt.

The gateway operates statelessly — rename channels, modify playlists, or update configurations without breaking existing
client links. Everything just continues to work.

When multiple clients request the same stream, the gateway automatically demultiplexes from a single upstream
connection, reducing bandwidth and server load.

**Need to transcode?** Configure any FFmpeg command or external tool for stream processing. The gateway handles the
pipeline, you define the transformation.

### :material-tune: Flexible Configuration

Build rules that cascade from global defaults down to individual client overrides. Apply configurations at any level:

- **Global** — Default behavior for your entire gateway
- **Playlist** — Specific settings per IPTV source
- **Preset** — Reusable client templates for common scenarios
- **Client** — Individual customization when needed

Mix and match rules to create exactly the experience each client requires.

<div class="grid cards" markdown>

- :material-cached:{ .lg .middle } [Intelligent Caching](config/cache.md)

    ---

    All remote content is cached on disk with proper header respect. The gateway reuses cached data whenever possible, reducing upstream requests.

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