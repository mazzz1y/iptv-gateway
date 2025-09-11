# Mark Hidden

The `mark_hidden` rule marks channels as hidden, removing them from metrics and logs while keeping them in the playlist.

## YAML Structure

```yaml
channel_rules:
  - mark_hidden:
      when:
        name_patterns: ["<regex>"]
        # or attr/tag/and/or/invert conditions
```

## Fields

| Field | Type               | Required | Description                                         |
|-------|--------------------|----------|-----------------------------------------------------|
| when  | [When](../when.md) | Yes      | Conditions specifying which channels to mark hidden |

## Examples

### Hide Test Channels from Monitoring

```yaml
channel_rules:
  - mark_hidden:
      when:
        name_patterns: ["(?i).*test.*"]
```
