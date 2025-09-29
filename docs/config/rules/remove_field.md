# Remove Field

The `remove_field` rule removes attributes/tags from channels, matching by selector and patterns.

## YAML Structure

```yaml
channel_rules:
  - remove_field:
      selector: {...} # what to remove (attribute/tag)
      patterns: [ ... ] # patterns to match (see below)
      condition: {...} # optional, see [Condition](when.md)
```

## Fields

| Field      | Type                       | Required | Description                   |
|------------|----------------------------|----------|-------------------------------|
| selector   | [`Selector`](../common.md) | Yes      | What to remove (attribute/tag)|
| patterns   | `[]regex`                   | Yes      | Patterns (regex) to match     |
| condition  | [`Condition`](condition.md)     | No       | Optional, restricts rule      |

## Examples

Remove all "tvg-logo" or "tvg-id" attributes from international channels:

```yaml
channel_rules:
  - remove_field:
      selector:
        attr: "*"
      patterns: ["tvg-logo", "tvg-id"]
      condition:
        selector:
          attr: "group-title"
        patterns: ["^International$"]
```

