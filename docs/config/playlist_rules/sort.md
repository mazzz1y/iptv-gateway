# Sort Playlist Rule

The `sort` rule allows you to control the order in which channels are presented in a playlist, with support for grouping
and custom ordering patterns.

## YAML Structure

```yaml
playlist_rules:
  - sort:
      attr: ""
      tag: ""
      order: []
      group_by:
        attr: ""
        tag: ""
        group_order: []
```

## Fields

| Field      | Type          | Required | Description                                                                |
|------------|---------------|----------|----------------------------------------------------------------------------|
| `attr`     | `string`      | No       | Attribute name to use for sorting channels (mutually exclusive with `tag`) |
| `tag`      | `string`      | No       | Tag name to use for sorting channels (mutually exclusive with `attr`)      |
| `order`    | `[]regex`     | No       | Custom order patterns for channels (regex patterns)                        |
| `group_by` | `GroupByRule` | No       | Group channels before sorting                                              |

### GroupByRule

| Field         | Type      | Required | Description                                       |
|---------------|-----------|----------|---------------------------------------------------|
| `attr`        | `string`  | No*      | Attribute name to group by                        |
| `tag`         | `string`  | No*      | Tag name to group by                              |
| `group_order` | `[]regex` | No       | Custom order patterns for groups (regex patterns) |

\* Either `attr` or `tag` must be specified when using `group_by`.

## How It Works

1. **Without grouping**: Channels are sorted based on their names or the specified `attr`/`tag` value
2. **With custom order**: The `order` array defines priority patterns (regex). Channels matching earlier patterns get
   higher priority
3. **With grouping**: Channels are first grouped by the `group_by` field, then:
    - Groups are sorted according to `group_order` patterns
    - Within each group, channels are sorted by their names or specified field
4. **Pattern matching**: Each pattern is treated as a regex. Empty strings (`""`) act as wildcards for unmatched items
5. **Fallback sorting**: Items with the same priority are sorted alphabetically

## Examples

### Basic Channel Sorting

```yaml
# Sort all channels alphabetically by name
playlist_rules:
  - sort: {}
```

### Custom Priority Order

```yaml
# Move Sports and Music channels to the end of the playlist. Everything else is sorted alphabetically.
playlist_rules:
  - sort:
      order: ["", "Sports.*", "Music.*"]
```

### Sort by Attribute Value

```yaml
# Sort channels by their tvg-name attribute
playlist_rules:
  - sort:
      attr: "tvg-name"
```

### Group by Category with Custom Group Order

```yaml
# Group channels by group-title, with News first, then Sports, then everything else
playlist_rules:
  - sort:
      group_by:
        attr: "group-title"
        group_order: ["News", "Sports", ""]
```

### Complex Example: Group and Sort with Patterns

```yaml
# Group by EXTGRP tag, prioritize HD channels within each group
playlist_rules:
  - sort:
      order: [".*HD.*", ".*FHD.*", ""]  # HD channels first in each group
      group_by:
        tag: "EXTGRP"
        group_order: ["Premium", "Movies", ""]  # Premium group first
```