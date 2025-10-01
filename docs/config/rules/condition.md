# Condition Blocks

The `condition` block controls when a rule is applied, based on channel properties, client, playlist, and more

!!! note
    All fields are optional. To combine criteria use `and` or `or`, which take arrays of condition blocks.

## YAML Structure

```yaml
condition:
  selector: ""
  patterns: []
  clients: []
  playlists: []
  and: []
  or: []
  invert: false
```

## Fields

| Field      | Type                           | Description                                                         |
|------------|--------------------------------|---------------------------------------------------------------------|
| `selector` | [`Selector`](./selector.md)    | See selector docs for details on matching properties                |
| `patterns` | `[]regex`                      | Array of regex patterns, matches channel name or other selector item |
| `clients`  | `[]string`                     | Restrict to clients by name                                         |
| `playlists`| `[]string`                     | Restrict to playlists by name                                       |
| `and`      | [`[]Condition`](./condition.md) | All nested conditions must match                                    |
| `or`       | [`[]Condition`](./condition.md) | At least one nested condition must match                            |
| `invert`   | `boolean`                      | If true, invert the condition result                                |


## Examples

Channel Name Pattern:
```yaml
condition:
  patterns: ["^CNN.*", "^BBC.*"]
```

Limit to Clients/Playlists:
```yaml
condition:
  clients: ["family-tablet", "living-room-tv"]
  playlists: ["sports-premium", "news-channels"]
```

Attribute Match Using Selector:
```yaml
condition:
  selector: "attr/group-title"
  patterns: ["^Sports$"]
```

Nested Conditions with AND/OR:
```yaml
condition:
  or:
    - clients: ["premium-client"]
    - patterns: ["^HD .*"]
```

Invert Condition:
```yaml
condition:
  patterns: ["^Music .*"]
  invert: true
```
