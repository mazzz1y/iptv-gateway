# Presets

The presets block represents a collection of reusable configuration templates. Presets allow you to define common sets of rules, proxy settings, and subscriptions that can be applied to multiple clients. This is particularly useful when you have similar configurations across different devices or users.

## YAML Structure

```yaml
presets:
  - name: preset-name
    channel_rules: []
    playlist_rules: []
    proxy: {}
    subscriptions: []
```

## Fields

| Field            | Type                | Required | Description                                          |
|------------------|---------------------|----------|------------------------------------------------------|
| `name`           | `string`            | Yes      | Unique name identifier for this preset              |
| `channel_rules`  | `[]rule`            | No       | Array of channel processing rules to apply          |
| `playlist_rules` | `[]rule`            | No       | Array of playlist processing rules to apply         |
| `proxy`          | [Proxy](./proxy.md) | No       | Proxy configuration settings                         |
| `subscriptions`  | `[]string`          | No       | List of subscription names to include in this preset |

## Examples

### Basic Quality Preset

```yaml
presets:
  - name: hd-quality
    playlist_rules:
      - remove_duplicates:
          - patterns: ["4K", "UHD", "FHD", "HD", ""]
```

### Family-Friendly Preset

```yaml
presets:
  - name: family-safe
    channel_rules:
      - remove_channel: {}
        when:
          - attr:
              name: "group-title"
              value: "(?i)(adult|xxx|18\\+)"
      - set_field:
          - attr:
              name: "group-title"
              template: "Family Safe"
        when:
          - name: ".*Kids.*"
    subscriptions: ["family-channels", "educational"]
```

### Performance Optimized Preset

```yaml
presets:
  - name: low-bandwidth
    proxy:
      enabled: true
      concurrency: 2
    channel_rules:
      - remove_channel: {}
        when:
          - name: ".*(4K|UHD).*"
      - set_field:
          - attr:
              name: "group-title"
              template: "SD Quality - {{ .Channel.Attrs.group-title }}"
    subscriptions: ["basic-package"]
```