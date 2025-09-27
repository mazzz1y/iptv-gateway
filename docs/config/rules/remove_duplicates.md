# Remove Duplicates

The `remove_duplicates` rule identifies and removes duplicate channels based on various criteria. This is useful for
cleaning up playlists that contain the same channel in multiple resolutions, quality levels, timezones, etc.

!!! note
This rule applies to the entire channel list after channel-level rules are processed.

## YAML Structure

```yaml
rules:
  - remove_duplicates:
      name_patterns: []
      attr:
        name: ""
        patterns: []
      tag:
        name: ""
        patterns: []
      set_field: ""
      when:
        clients: []
```

## Fields

| Field           | Type                           | Required    | Description                                                     |
|-----------------|--------------------------------|-------------|-----------------------------------------------------------------|
| `name_patterns` | `[]regex`                      | Conditional | Name patterns ordered by priority (highest first)               |
| `attr`          | [`NamePatterns`](../common.md) | Conditional | Match duplicates by attribute using `name` and `patterns` array |
| `tag`           | [`NamePatterns`](../common.md) | Conditional | Match duplicates by tag using `name` and `patterns` array       |
| `set_field`     | `template`                     | Optional    | Template for the final name                                     |
| `when`          | [When](when.md)                | Optional    | Conditions specifying when to apply (only `clients` supported)  |

*Exactly one of `name_patterns`, `attr`, or `tag` must be specified.*

## How It Works

1. The rule scans channel names for configured patterns
2. Channels with the same base name (after removing patterns) are grouped as duplicates
3. The rule only processes groups that contain channels with different extractable patterns
4. Among pattern-matching duplicates, the channel with the highest priority pattern is kept
5. If `set_field` is provided, it is appended to the base name of the selected channel
6. Channels with identical names but no matching patterns are left untouched

## Examples

### Basic Quality-Based Deduplication

```yaml
# Input: CNN, CNN HD, CNN 4K, ESPN, ESPN FHD, ESPN UHD, Fox News
# Output: CNN 4K, ESPN UHD, Fox News
rules:
  - remove_duplicates:
      name_patterns: ["4K", "UHD", "FHD", "HD", ""]
```

### With Custom Template

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K, National Geographic UHD, National Geographic
# Output: Discovery Channel HQ-Preferred, National Geographic HQ-Preferred
rules:
  - remove_duplicates:
      name_patterns: ["4K", "UHD", "FHD", "HD", ""]
      set_field: "{{.BaseName}} HQ-Preferred"
```

### Trimming Patterns

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K
# Output: Discovery Channel (patterns removed)
rules:
  - remove_duplicates:
      name_patterns: ["4K", "UHD", "FHD", "HD", ""]
      set_field: "{{.BaseName}}"
```

## Template

List of available template variables:

- `{{.BaseName}}` - Base name with patterns removed
- `{{.Channel.Name}}` - Best channel name
- `{{.Channel.Attrs.key}}` - Best channel attributes map
- `{{.Channel.Tags.key}}` - Best channel tags map
