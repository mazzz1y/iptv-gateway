# Merge Channels

The `merge_channels` rule combines multiple channel variants into a single channel. Duplicate variants act as fallbacks
if the primary channel becomes unavailable (for example, due to upstream issues or proxy limits).

!!! note
This rule applies to the entire channel list after channel-level rules are processed.

## YAML Structure

```yaml
rules:
  - merge_channels:
      name_patterns: []
      attr:
        name: "tvg-id"
        patterns: []
      set_field: ""
      when:
        clients: []
```

## Fields

| Field           | Type                           | Required    | Description                                                    |
|-----------------|--------------------------------|-------------|----------------------------------------------------------------|
| `name_patterns` | `[]regex`                      | Conditional | Name patterns to identify channels for merging                 |
| `attr`          | [`NamePatterns`](../common.md) | Conditional | Match channels by `tvg-id` attribute only                      |
| `set_field`     | `template`                     | Optional    | Template for the final name                                    |
| `when`          | [When](when.md)                | Optional    | Conditions specifying when to apply (only `clients` supported) |

*Exactly one of `name_patterns` or `attr` must be specified. Only `tvg-id` attribute is allowed.*

## How It Works

1. The rule scans all channels for configured patterns
2. Any channel matching a pattern is processed
3. If `set_field` is specified, it transforms the channel name accordingly
4. Multiple channels with identical names will be collapsed by the application
5. Channels without matching patterns remain unchanged

## Examples

### Basic Channel Merging

```yaml
# Input: CNN HD, CNN 4K, ESPN UHD, Fox News
# Output: CNN Multi-Quality, ESPN UHD, Fox News
rules:
  - merge_channels:
      name_patterns: ["4K", "UHD", "FHD", "HD"]
      set_field: "{{.BaseName}} Multi-Quality"
```

### Trimming Patterns

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K
# Output: Discovery Channel
rules:
  - merge_channels:
      name_patterns: ["4K", "UHD", "FHD", "HD"]
      set_field: "{{.BaseName}}"
```

### Using Best Channel Name

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K
# Output: Discovery Channel 4K
rules:
  - merge_channels:
      name_patterns: ["4K", "HD"]
```

### Attribute-Based Merging

```yaml
# Merge channels based on tvg-id patterns
rules:
  - merge_channels:
      attr:
        name: "tvg-id"
        patterns: ["HD", "4K", "UHD"]
      set_field: "{{.BaseName}} Multi-Source"
```

## Template

List of available template variables:

- `{{.BaseName}}` - Base name with patterns removed (e.g., "CNN" from "CNN HD")
- `{{.Channel.Name}}` - Best channel name (e.g., "CNN HD")
- `{{.Channel.Attrs}}` - Best channel attributes map
- `{{.Channel.Tags}}` - Best channel tags map
