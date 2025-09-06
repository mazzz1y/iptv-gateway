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

| Field            | Type                                             | Description                                               |
|------------------|--------------------------------------------------|-----------------------------------------------------------|
| `listen_addr`    | `string`                                         | Server listening address                                  |
| `metrics_addr`   | `string`                                         | Prometheus metrics server address                         |
| `public_url`     | `string`                                         | Public URL for generating links                           |
| `url_generator`  | [URL Generator](./config/url_generator.md)       | URL generation and encryption configuration               |
| `log`            | [Log](./config/log.md)                           | Logging configuration                                     |
| `proxy`          | [Proxy](./config/proxy.md)                       | Stream proxy configuration for remuxing with ffmpeg       |
| `cache`          | [Cache](./config/cache.md)                       | Cache configuration for playlists and EPGs                |
| `playlists`      | [Playlists](./config/playlists.md)               | Array of playlist definitions with sources and rules      |
| `epgs`           | [EPGs](./config/epgs.md)                         | Array of EPG definitions with sources                     |
| `conditions`     | [Conditions](./config/conditions.md)             | Named condition definitions for reuse in rules            |
| `channel_rules`  | [Channel Rules](config/channel_rules/index.md)   | Global channel processing rules                           |
| `playlist_rules` | [Playlist Rules](config/playlist_rules/index.md) | Global playlist processing rules                          |
| `presets`        | [Presets](./config/presets.md)                   | Array of reusable configuration templates                 |
| `clients`        | [Clients](./config/clients.md)                   | Array of IPTV client definitions with individual settings |

## Example Configuration

!!! note "Client Links"

    Each client can access the following endpoints:

    - `{public_url}/{client_secret}/playlist.m3u8`
    - `{public_url}/{client_secret}/epg.xml`
    - `{public_url}/{client_secret}/epg.xml.gz`

```yaml
listen_addr: ":8080"
metrics_addr: ":9090"
public_url: "https://iptv.example.com"

log:
  level: info
  format: text

proxy:
  enabled: true
  concurrency: 10

playlists:
  - name: main-playlist
    source: "http://example.com/playlist.m3u8"

epgs:
  - name: main-epg
    source: "http://example.com/epg.xml.gz"

conditions:
  - name: "adult"
    when:
      - attr:
          name: "group-title"
          value: "(?i)adult"

channel_rules:
  - remove_channel: true
    when: adult

playlist_rules:
  - remove_duplicates:
      - patterns: ["4K", "UHD", "FHD", "HD", ""]

presets:
  - name: family-friendly
    channel_rules:
      - remove_channel: true
        when: adult

clients:
  - name: living-room-tv
    secret: "living-room-tv-secret-789"
    preset: "family-friendly"
    playlist: "main-playlist"
    epg: "main-epg"
```