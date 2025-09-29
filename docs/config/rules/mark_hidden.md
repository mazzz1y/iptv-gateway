# Mark Hidden

The `mark_hidden` rule marks channels as hidden for all clients matching the given `condition`.

## YAML Structure

```yaml
channel_rules:
  - mark_hidden:
      condition: { ... } # See [Condition](when.md)
```

## Fields

| Field      | Type                   | Required | Description                          |
|------------|------------------------|----------|--------------------------------------|
| condition  | [`Condition`](condition.md) | Yes      | Which channels will be hidden        |

## Example

```yaml
channel_rules:
  - mark_hidden:
      condition:
        clients: ["test-client"]
```
