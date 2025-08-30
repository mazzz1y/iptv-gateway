# Remove Channel Duplicates

The `remove_channel_dups` rule identifies and removes duplicate channels based on quality patterns in channel names, keeping only the highest quality version. This is essential for cleaning up playlists that contain the same channel in multiple resolutions or quality levels.

## YAML Structure

```yaml
remove_channel_dups:
  - patterns: []
    trim_pattern: false
```

## Fields

| Field          | Type       | Required | Description                                             |
|----------------|------------|----------|---------------------------------------------------------|
| `patterns`     | `[]string` | Yes      | Quality patterns ordered by priority (highest first)   |
| `trim_pattern` | `bool`     | No       | Remove the quality pattern from final channel name     |

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

### Sports Channels Cleanup

```yaml
# Prioritize higher quality sports channels
remove_channel_dups:
  - patterns: ["4K", "UHD", "2160p", "FHD", "1080p", "HD", "720p", "SD", ""]
    trim_pattern: false
```

### Multiple Pattern Configurations

```yaml
remove_channel_dups:
  # First pass: Remove quality duplicates
  - patterns: ["4K", "UHD", "FHD", "HD", ""]
    trim_pattern: true
  # Second pass: Remove language duplicates  
  - patterns: ["ENG", "EN", ""]
    trim_pattern: true
```

### Conditional Deduplication

```yaml
# Only apply to specific channel groups
remove_channel_dups:
  - patterns: ["4K", "UHD", "FHD", "HD", "SD", ""]
    trim_pattern: true
when:
  - attr:
      name: "group-title"
      value: "(Sports|Movies|Entertainment)"
```