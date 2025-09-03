# Playlists

Playlists define collections of IPTV channels from M3U/M3U8 sources. Each playlist can contain multiple sources and custom processing rules.

## YAML Structure

```yaml
playlists:
  - name: playlist-name
    sources:
      - "http://example.com/playlist-1.m3u8"
      - "/path/to/local/playlist-2.m3u8"
    proxy:
      enabled: true
      concurrency: 2
    channel_rules:
      - remove_channel: { }
        when:
          - name: "^Test.*"
```

## Fields

| Field            | Type                                         | Required | Description                                                       |
|------------------|----------------------------------------------|----------|-------------------------------------------------------------------|
| `name`           | `string`                                     | Yes      | Unique name identifier for this playlist                          |
| `sources`        | `[]string`                                   | Yes      | Array of playlist sources (URLs or file paths, M3U/M3U8 format)  |
| `proxy`          | [Proxy](./proxy.md)                          | No       | Playlist-specific proxy configuration                             |
| `concurrency`    | `int`                                        | No       | Concurrency for downloading/processing sources                    |
| `channel_rules`  | [[]Channel Rule](./channel_rules/index.md)  | No       | Array of channel processing rules applied to this playlist       |
| `playlist_rules` | [[]Playlist Rule](./playlist_rules/index.md) | No       | Array of playlist processing rules applied to this playlist      |

## Examples

### Basic Playlist

```yaml
playlists:
  - name: basic-tv
    sources:
      - "https://provider.com/basic.m3u8"
```

### Sports Package

```yaml
playlists:
  - name: sports-premium
    sources:
      - "https://sports-provider.com/premium.m3u8"
      - "https://sports-provider.com/international.m3u8"
    playlist_rules:
      - remove_duplicates:
          - patterns: [ "4K", "UHD", "FHD", "HD", "" ]
    channel_rules:
      - set_field:
          - attr:
              name: "group-title"
              template: "Sports"
        when:
          - name: ".*ESPN.*|.*Fox Sports.*|.*Sky Sports.*"
```

### Family Safe Playlist

```yaml
playlists:
  - name: family-safe
    sources:
      - "https://family-provider.com/channels.m3u8"
    proxy:
      enabled: true
      concurrency: 5
    channel_rules:
      - remove_channel: {}
        when:
          - or:
              - attr:
                  name: "group-title"
                  value: "(?i)(adult|xxx|18+)"
              - name: "(?i).*(adult|xxx|mature).*"
```
