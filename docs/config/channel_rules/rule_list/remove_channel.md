# Remove Channel

The `remove_channel` rule completely removes channels from the playlist based on specified conditions. This rule is
particularly useful for content filtering, removing test channels, or excluding unwanted content categories.

## YAML Structure

```yaml
remove_channel: true
```

## Fields

This rule does not accept any configuration fields. It is designed to be used exclusively with the `when` clause to
specify which channels should be removed.

| Field  | Type | Required | Description                          |
|--------|------|----------|--------------------------------------|
| (none) | -    | -        | This rule has no configurable fields |

## Examples

### Remove Adult Content

```yaml
remove_channel: {}
when:
  - attr:
      name: "group-title"
      value: "(?i)(adult|xxx|18\\+)"
```