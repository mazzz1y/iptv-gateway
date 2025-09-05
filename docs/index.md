# IPTV Gateway

> ⚠️ **Warning**: This project is currently under construction. Some features might be unstable or incomplete.

A minimal, functional IPTV Gateway for your TV.

## Features

- **Stream Proxying**: M3U8 and EPG proxying with stream remuxing support
- **Playlist Management**: Merging playlists and EPGs
- **Playlist Transformations**: Advanced transformations based on M3U8 tags, attributes, or channel names
- **Flexible Rules**: Global, per-subscription, per-preset, and per-client rule configuration
- **Caching**: On-disk caching
- **Error Handling**: Custom error screens with configurable streaming commands
- **Rate Limiting**: Multi-level rate limiting (global, subscription, client)
- **Low Memory Usage**: Efficient streaming with on-the-fly processing
- **Monitoring**: Prometheus metrics support

## Typical Use Cases

- Set up this app as a single point of access for your IPTV subscription
- Assign different playlists to different clients
- Reorganize channels
- Merge multiple EPGs to complete your TV guide
- Filter channels for your kids
- Set up a custom error page for your clients
- View watch stats in Grafana
- And much more...