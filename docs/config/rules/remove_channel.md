# Remove Channel

The `remove_channel` rule deletes entire channels matching a `condition`.

## YAML Structure

```yaml
channel_rules:
  - remove_channel:
      condition: { ... } # See [Condition](when.md)
```

## Fields

| Field      | Type                       | Required | Description            |
|------------|----------------------------|----------|------------------------|
| condition  | [`Condition`](condition.md)     | Yes      | Which channels to drop |


## Example

Remove all music channels:

```yaml
channel_rules:
  - remove_channel:
      condition:
        patterns: ["^Music .*"]
```
