# Condition Blocks

The `condition` block (previously `when`) controls *when* a rule is applied, based on channel properties, client, playlist, and more. All fields are optional—absent fields mean "match all".

## YAML Structure

```yaml
condition:
  selector:
    # See Selector documentation for available options
  patterns: [ ... ]   # Regex patterns to match
  clients: [ ... ]    # Apply only for these clients
  playlists: [ ... ]  # Apply only for these playlists
  and:
    - { ... }         # Nested conditions (all must match)
  or:
    - { ... }         # Nested conditions (any must match)
  invert: false       # Invert this condition (optional)
```

> All fields are optional. To combine criteria use `and` or `or`, which take arrays of condition blocks.

## Fields

| Field      | Type                          | Description                                                         |
|------------|------------------------------|---------------------------------------------------------------------|
| `selector` | [`Selector`](../common.md)   | See selector docs for details on matching properties                |
| `patterns` | `[]regex`                    | Array of regex patterns, matches channel name or other selector item |
| `clients`  | `[]string`                   | Restrict to clients by name                                         |
| `playlists`| `[]string`                   | Restrict to playlists by name                                       |
| `and`      | [`[]Condition`](condition.md)| All nested conditions must match                                    |
| `or`       | [`[]Condition`](condition.md)| At least one nested condition must match                            |
| `invert`   | `boolean`                    | If true, invert the condition result                                |


## Examples

### Channel Name Pattern
```yaml
condition:
  patterns: ["^CNN.*", "^BBC.*"]
```

### Limit to Clients/Playlists
```yaml
condition:
  clients: ["family-tablet", "living-room-tv"]
  playlists: ["sports-premium", "news-channels"]
```

### Attribute Match Using Selector
```yaml
condition:
  selector:
    attr: "group-title"
  patterns: ["^Sports$"]
```

### Nested Conditions with AND/OR
```yaml
condition:
  and:
    - clients: ["premium-client"]
    - patterns: ["^HD .*"]
```

### Invert Condition
```yaml
condition:
  patterns: ["^Music .*"]
  invert: true
```
