# Presets

The presets block represents a collection of reusable configuration templates. Presets allow you to define common sets
of rules, proxy settings, and subscriptions that can be applied to multiple clients. This is particularly useful when
you have similar configurations across different devices or users.

## YAML Structure

```yaml
presets:
  - name: ""
    playlists: []
    epgs: []
    channel_rules: []
    playlist_rules: []
    proxy: {}
```

## Fields

| Field            | Type                                         | Required | Description                                 |
|------------------|----------------------------------------------|----------|---------------------------------------------|
| `name`           | `string`                                     | Yes      | Unique name identifier for this preset      |
| `playlists`      | `[]string`                                   | No       | Playlist name(s) to include in this preset  |
| `epgs`           | `[]string`                                   | No       | EPG name(s) to include in this preset       |
| `channel_rules`  | [[]Channel Rule](./channel_rules/index.md)   | No       | Array of channel processing rules to apply  |
| `playlist_rules` | [[]Playlist Rule](./playlist_rules/index.md) | No       | Array of playlist processing rules to apply |
| `proxy`          | [Proxy](./proxy.md)                          | No       | Proxy configuration settings                |

## Examples

### Basic Quality Preset

```yaml
presets:
  - name: hd-quality
    playlist_rules:
      - remove_duplicates:
          name_patterns: ["4K", "UHD", "FHD", "HD", ""]
```

### Sports Package Preset

```yaml
presets:
  - name: sports-package
    playlists: ["sports-premium"]
    epgs: ["sports-guide"]
    playlist_rules:
      - remove_duplicates:
          name_patterns: ["4K", "UHD", "FHD", "HD", ""]
    channel_rules:
      - set_field:
          attr:
            name: "group-title"
            patterns: ["Sports"]
          when:
            name_patterns: [".*ESPN.*|.*Fox Sports.*|.*Sky Sports.*"]
```

### Family-Friendly Preset

```yaml
presets:
  - name: family-safe
    playlists: ["family-channels", "educational"]
    epgs: ["family-guide"]
    channel_rules:
      - remove_channel:
          when:
            attr:
              name: "group-title"
              patterns: ["(?i)(adult|xxx|18\+)"]
      - set_field:
          attr:
            name: "group-title"
            patterns: ["Family Safe"]
          when:
            name_patterns: ".*Kids.*"
```

### Complete Preset with Playlists and EPGs

```yaml
presets:
  - name: entertainment-package
    playlists: ["basic-tv", "movies", "series"]
    epgs: ["tv-guide", "international-guide"]
    proxy:
      enabled: true
    playlist_rules:
      - action: remove_duplicates
        name_patterns: ["4K", "UHD", "FHD", "HD", ""]
    channel_rules:
      - action: set_field
        attr:
          name: "group-title"
          template: "Entertainment - {{ .Channel.Attrs.group-title }}"
      - action: remove_channel
        when:
          attr:
            name: "group-title"
            patterns: ["(?i)(news|sports)"]
```
