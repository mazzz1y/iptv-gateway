# Remove Channel

The `remove_channel` rule completely removes channels from the playlist based on specified conditions. Use this rule
when you want to exclude channels entirely.

## YAML Structure

```yaml
channel_rules:
  - remove_channel:
      when:
        name_patterns: ["<regex>"]
        # or attr/tag/and/or/invert conditions
```

## Fields

| Field | Type               | Required | Description                                    |
|-------|------------------- |----------|------------------------------------------------|
| when  | [When](when.md) | Yes      | Conditions specifying which channels to remove |

## Examples

### Remove Adult Content

```yaml
channel_rules:
  - remove_channel:
      when:
        attr:
          name: "group-title"
          patterns: ["(?i)(adult|xxx|18\+)"]
```
