# Remove Duplicates

The `remove_duplicates` rule identifies and removes duplicate channels based on patterns in channel names.
This is useful for cleaning up playlists that contain the same channel in multiple resolutions, quality levels,
timezones, etc.

## YAML Structure

```yaml
remove_duplicates:
  - patterns: []
    trim_pattern: false
```

## Fields

| Field          | Type      | Required | Description                                  |
|----------------|-----------|----------|----------------------------------------------|
| `patterns`     | `[]regex` | Yes      | Patterns ordered by priority (highest first) |
| `trim_pattern` | `bool`    | No       | Remove the pattern from final channel name   |

## How It Works

1. **Pattern Matching**: The rule scans channel names for quality patterns
2. **Grouping**: Channels with the same base name (after removing patterns) are grouped as duplicates
3. **Priority Selection**: Among duplicates, the channel with the highest priority pattern is kept
4. **Pattern Order**: The first pattern in the array has the highest priority
5. **Trimming**: If `trim_pattern` is true, the quality pattern is removed from the final name

## Examples

### Basic Quality-Based Deduplication

```yaml
# Input: CNN, CNN HD, CNN 4K, ESPN, ESPN FHD, ESPN UHD, Fox News
# Output: CNN 4K, ESPN UHD, Fox News
remove_channel_dups:
  - patterns: ["4K", "UHD", "FHD", "HD", ""]
```

### With Pattern Trimming Enabled

```yaml
# Input: Discovery Channel HD, Discovery Channel 4K, National Geographic UHD, National Geographic
# Output: Discovery Channel, National Geographic
remove_channel_dups:
  - patterns: ["4K", "UHD", "FHD", "HD", ""]
    trim_pattern: true
```

### Multiple Pattern Configurations

```yaml
remove_channel_dups:
  # First pass: Remove quality duplicates
  - patterns: ["4K", "UHD", "FHD", "HD", ""]
    trim_pattern: true
  # Second pass: Remove language duplicates
  - patterns: ["EN", "DE", ""]
    trim_pattern: true
```