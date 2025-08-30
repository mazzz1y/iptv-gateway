# Subscriptions

Subscriptions define collections of IPTV playlists and Electronic Program Guides (EPGs) that clients can access. Each subscription can contain multiple playlist sources, EPG sources, custom proxy settings, and processing rules specific to that subscription's content.

## YAML Structure

```yaml
subscriptions:
  subscription_name:
    playlists:
      - "http://example.com/playlist-1.m3u8"
      - "http://example.com/playlist-2.m3u8"
    epgs:
      - "http://example.com/epg-1.xml"
      - "http://example.com/epg-2.xml.gz"
    proxy:
      enabled: true
      concurrency: 2
    rules:
      - remove_channel: {}
        when:
          - name: "^Test.*"
```

## Fields

| Field       | Type       | Required | Description                                           |
|-------------|------------|----------|-------------------------------------------------------|
| `playlists` | `[]string` | No       | Array of playlist URLs (M3U/M3U8 format)             |
| `epgs`      | `[]string` | No       | Array of EPG URLs (XML format, .gz compressed files supported) |
| `proxy`     | `object`   | No       | Subscription-specific proxy configuration             |
| `rules`     | `[]rule`   | No       | Array of processing rules applied to this subscription |

## Examples

### Basic Subscription

```yaml
subscriptions:
  basic-tv:
    playlists:
      - "https://provider.com/basic.m3u8"
    epgs:
      - "https://provider.com/guide.xml"
```

### Sports Package with Quality Control

```yaml
subscriptions:
  sports-premium:
    playlists:
      - "https://sports-provider.com/premium.m3u8"
      - "https://sports-provider.com/international.m3u8"
    epgs:
      - "https://sports-provider.com/epg.xml.gz"
    rules:
      - remove_channel_dups:
          - patterns: ["4K", "UHD", "FHD", "HD", ""]
            trim_pattern: true
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
  family-safe:
    playlists:
      - "https://family-provider.com/channels.m3u8"
    epgs:
      - "https://family-provider.com/schedule.xml"
    proxy:
      enabled: true
      concurrency: 5
    rules:
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
  global-news:
    playlists:
      - "https://news-source-1.com/channels.m3u8"
      - "https://news-source-2.com/international.m3u8"
      - "https://news-source-3.com/local.m3u8"
    epgs:
      - "https://news-source-1.com/guide.xml"
      - "https://news-source-2.com/schedule.xml"
    rules:
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