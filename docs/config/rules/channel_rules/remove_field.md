# Remove Field

The `remove_field` rule removes attributes/tags from channels, matching by selector and patterns.

## YAML Structure

```yaml
remove_field:
  selector: {}
  condition: {}
```

## Fields

| Field     | Type                           | Required | Description                    |
|-----------|--------------------------------|----------|--------------------------------|
| selector  | [`Selector`](../selector.md)   | Yes      | What to remove (attribute/tag) |
| patterns  | `[]regex`                      | Yes      | Patterns (regex) to match      |
| condition | [`Condition`](../condition.md) | No       | Optional, restricts rule       |

## Examples

Remove "tvg-logo" attribute from international channels:

```yaml
channel_rules:
  - remove_field:
      selector: attr/tvg-logo
      condition:
        selector: attr/group-title
        patterns: ["^International$"]
```

