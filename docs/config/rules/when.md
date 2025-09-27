# When Conditions

The `when` clause is a conditional system used to apply rules only to specific channels that match certain criteria.

## YAML Structure

```yaml
when:
  name_patterns: [""]
  attr:
    name: ""
    patterns: [""]
  tag:
    name: ""
    patterns: [""]
  user: [""]
  playlist: [""]
  and: []
  or: []
  invert: false
```

## Fields

### Condition Fields

| Field           | Type                           | Required | Description                      |
|-----------------|--------------------------------|----------|----------------------------------|
| `name_patterns` | `[]regex`                      | No       | Match against channel name       |
| `attr`          | [`NamePatterns`](../common.md) | No       | Match against M3U attributes     |
| `tag`           | [`NamePatterns`](../common.md) | No       | Match against M3U tags           |
| `user`          | `[]string`                     | No       | Match against client names       |
| `playlist`      | `[]string`                     | No       | Match against playlist names     |
| `and`           | [`[]When`](./when.md)          | No       | All nested conditions must match |
| `or`            | [`[]When`](./when.md)          | No       | Any nested condition must match  |
| `invert`        | `bool`                         | No       | Invert result                    |

## Examples

### Basic Name Matching

```yaml
when:
  name_patterns: ["^CNN.*"]
```

### Attribute Matching

```yaml
when:
  attr:
    name: "group-title"
    patterns: ["^Sports$"]
```

### Tag Matching

```yaml
when:
  tag:
    name: "EXTGRP"
    patterns: ["^Entertainment$"]
```

### User-Specific Rule

```yaml
when:
  user: ["family-tablet", "living-room-tv"]
```

### Playlist-Specific Rule

```yaml
when:
  playlist: ["sports-premium", "news-channels"]
```

### Combined Conditions

```yaml
when:
  and:
    - user: ["premium-client"]
    - name_patterns: ["^HD .*"]
```

### Invert Condition

```yaml
when:
  name_patterns: ["^Music .*"]
  invert: true
```
