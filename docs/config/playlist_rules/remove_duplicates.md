# Remove Duplicates

The `remove_duplicates` rule identifies and removes duplicate channels based on various criteria. This is useful for
cleaning up playlists that contain the same channel in multiple resolutions, quality levels, timezones, etc.

## YAML Structure

```yaml
playlist_rules:
  - remove_duplicates:
      name_patterns: []
      attr:
        name: ""
        patterns: []
      tag:
        name: ""
        patterns: []
```

## Fields

| Field           | Type                           | Required    | Description                                                     |
|-----------------|--------------------------------|-------------|-----------------------------------------------------------------|
| `name_patterns` | `[]regex`                      | Conditional | Name patterns ordered by priority (highest first)               |
| `attr`          | [`NamePatterns`](../common.md) | Conditional | Match duplicates by attribute using `name` and `patterns` array |
| `tag`           | [`NamePatterns`](../common.md) | Conditional | Match duplicates by tag using `name` and `patterns` array       |

*Exactly one of `name_patterns`, `attr`, or `tag` must be specified.*

## How It Works

1. The rule scans channel names for quality patterns
2. Channels with the same base name (after removing patterns) are grouped as duplicates
3. Among duplicates, the channel with the highest priority pattern is kept
4. If `trim_pattern` is true, the quality pattern is removed from the final name

## Examples

### Basic Quality-Based Deduplication

```yaml
# Input: CNN, CNN HD, CNN 4K, ESPN, ESPN FHD, ESPN UHD, Fox News
# Output: CNN 4K, ESPN UHD, Fox News
playlist_rules:
  - remove_duplicates:
      name_patterns: ["4K", "UHD", "FHD", "HD", ""]
```

### With Pattern Trimming Enabled

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K, National Geographic UHD, National Geographic
# Output: Discovery Channel, National Geographic
playlist_rules:
  - remove_duplicates:
      name_patterns: ["4K", "UHD", "FHD", "HD", ""]
      trim_pattern: true
```

