# Presets

The presets block represents a collection of reusable configuration templates. Presets allow you to define common sets of rules, proxy settings, and subscriptions that can be applied to multiple clients. This is particularly useful when you have similar configurations across different devices or users.

## YAML Structure

```yaml
presets:
  preset_name:
    rules: []
    proxy: {}
    subscriptions: []
```

## Fields

| Field           | Type                | Required | Description                                          |
|-----------------|---------------------|----------|------------------------------------------------------|
| `rules`         | `[]rule`            | No       | Array of processing rules to apply                   |
| `proxy`         | [Proxy](./proxy.md) | No       | Proxy configuration settings                         |
| `subscriptions` | `[]string`          | No       | List of subscription names to include in this preset |

## Examples

### Basic Quality Preset

```yaml
presets:
  hd-quality:
    rules:
      - remove_channel_dups:
          - patterns: ["4K", "UHD", "FHD", "HD", ""]
            trim_pattern: true
```

### Family-Friendly Preset

```yaml
presets:
  family-safe:
    rules:
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
  low-bandwidth:
    proxy:
      enabled: true
      concurrency: 2
    rules:
      - remove_channel: {}
        when:
          - name: ".*(4K|UHD).*"
      - set_field:
          - attr:
              name: "group-title"
              template: "SD Quality - {{ .Channel.Attrs.group-title }}"
    subscriptions: ["basic-package"]
```