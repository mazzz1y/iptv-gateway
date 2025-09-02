# Subscriptions

Subscriptions define collections of IPTV playlists and Electronic Program Guides (EPGs) that clients can access. Each
subscription can contain multiple playlist sources, EPG sources, custom proxy settings, and processing rules specific to
that subscription's content.

## YAML Structure

```yaml
subscriptions:
  - name: subscription-name
    playlist_sources:
      - "http://example.com/playlist-1.m3u8"
      - "/path/to/local/playlist-2.m3u8"
    epg_sources:
      - "http://example.com/epg-1.xml"
      - "/path/to/local/epg-2.xml.gz"
    proxy:
      enabled: true
      concurrency: 2
    channel_rules:
      - remove_channel: { }
        when:
          - name: "^Test.*"
```

## Fields

| Field              | Type                                         | Required | Description                                                                           |
|--------------------|----------------------------------------------|----------|---------------------------------------------------------------------------------------|
| `name`             | `string`                                     | Yes      | Unique name identifier for this subscription                                          |
| `playlist_sources` | `[]string`                                   | No       | Array of playlist sources (URLs or file paths, M3U/M3U8 format)                       |
| `epg_sources`      | `[]string`                                   | No       | Array of EPG sources (URLs or file paths, XML format, .gz compressed files supported) |
| `proxy`            | [Proxy](./proxy.md)                          | No       | Subscription-specific proxy configuration                                             |
| `channel_rules`    | [[]Channel Rule](./channel_rules/index.md) | No       | Array of channel processing rules applied to this subscription                        |
| `playlist_rules`   | [[]Playlist Rule](./playlist_rules/index.md) | No       | Array of playlist processing rules applied to this subscription                       |

## Examples

### Basic Subscription

```yaml
subscriptions:
  - name: basic-tv
    playlist_sources:
      - "https://provider.com/basic.m3u8"
    epg_sources:
      - "https://provider.com/guide.xml"
```

### Sports Package with Quality Control

```yaml
subscriptions:
  - name: sports-premium
    playlist_sources:
      - "https://sports-provider.com/premium.m3u8"
      - "https://sports-provider.com/international.m3u8"
    epg_sources:
      - "https://sports-provider.com/epg.xml.gz"
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

### Family Subscription with Content Filtering

```yaml
subscriptions:
  - name: family-safe
    playlist_sources:
      - "https://family-provider.com/channels.m3u8"
    epg_sources:
      - "https://family-provider.com/schedule.xml"
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
      - set_field:
          - attr:
              name: "group-title"
              template: "Family - {{ .Channel.Attrs.group-title }}"
```

### Multi-Source News Subscription

```yaml
subscriptions:
  - name: global-news
    playlist_sources:
      - "https://news-source-1.com/channels.m3u8"
      - "https://news-source-2.com/international.m3u8"
      - "https://news-source-3.com/local.m3u8"
    epg_sources:
      - "https://news-source-1.com/guide.xml"
      - "https://news-source-2.com/schedule.xml"
    channel_rules:
      - set_field:
          - attr:
              name: "group-title"
              template: "News & Current Affairs"
        when:
          - name: ".*(News|BBC|CNN|Fox News|MSNBC).*"
      - remove_channel: {}
        when:
          - name: ".*Test.*|.*Sample.*"
```

