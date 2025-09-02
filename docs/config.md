# Configuration

IPTV Gateway can read configuration from a file or from a directory by combining multiple files based on
top-level elements.

By default, it reads configuration from `config.yaml` in the current directory.

```bash
iptv-gateway -config ./config.yaml # from file
iptv-gateway -config ./config      # from directory
```

!!! note "Hint"

    All arrays with a single value can be specified without brackets.

### Root Level Configuration

| Field            | Type                                                   | Description                                                       |
|------------------|--------------------------------------------------------|-------------------------------------------------------------------|
| `listen_addr`    | `string`                                               | Server listening address                                          |
| `metrics_addr`   | `string`                                               | Prometheus metrics server address                                 |
| `public_url`     | `string`                                               | Public URL for generating links                                   |
| `log_level`      | `string`                                               | Logging level (debug, info, warn, error)                          |
| `secret`         | `string`                                               | Secret used as for encryption purposes                            |
| `proxy`          | [Proxy](./config/proxy.md)                             | Stream proxy configuration for remuxing with ffmpeg               |
| `cache`          | [Cache](./config/cache.md)                             | Cache configuration for playlists and EPGs                        |
| `subscriptions`  | [Subscriptions](./config/subscriptions.md)             | Array of subscription definitions with playlists, EPGs, and rules |
| `channel_rules`  | [Channel Rules](config/channel_rules/channel_rules.md) | Global channel processing rules                                   |
| `playlist_rules` | [Playlist Rules](config/playlist_rules/index.md)       | Global playlist processing rules                                  |
| `presets`        | [Presets](./config/presets.md)                         | Array of reusable configuration templates                         |
| `clients`        | [Clients](./config/clients.md)                         | Array of IPTV client definitions with individual settings         |

## Example Configuration

```yaml
# https://iptv.example.com/tv-secret/playlist.m3u8
# https://iptv.example.com/tv-secret/epg.xml
# https://iptv.example.com/tv-secret/epg.xml.gz

listen_addr: ":8080"l
metrics_addr: ":9090"
public_url: "https://iptv.example.com"
secret: "global-secret"

proxy:
  enabled: true
  concurrency: 10

subscriptions:
  - name: main-subscription
    playlist_sources: "http://example.com/playlist.m3u8"
    epg_sources: "http://example.com/epg.xml.gz"

channel_rules:
  - remove_channel: {}
    when:
      - attr:
          name: "group-title"
          value: "(?i)adult"

playlist_rules:
  - remove_duplicates:
      - patterns: [ "4K", "UHD", "FHD", "HD", "" ]

presets:
  - name: family-friendly
    channel_rules:
      - remove_channel: {}
        when:
          - attr:
              name: "group-title"
              value: "(?i)adult"

clients:
  - name: living-room-tv
    secret: "tv-secret"
    presets: "family-friendly"
    subscriptions: "main-subscription"
```