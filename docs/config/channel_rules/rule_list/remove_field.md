# Remove Field

The `remove_field` rule deletes specific channel metadata.

## YAML Structure

```yaml
channel_rules:
  - remove_field:
      tag_patterns: ["regex1", "regex2"] # Remove tags matching patterns
      attr_patterns: ["regex1", "regex2"]# Remove attributes matching patterns
      when:
        name_patterns: ["<regex>"]
        # or attr/tag/and/or/invert conditions
```

## Fields

| Field         | Type               | Required     | Description                            |
|---------------|--------------------|--------------|----------------------------------------|
| tag_patterns  | `[]regex`          | Conditional* | Regex patterns of tags to remove       |
| attr_patterns | `[]regex`          | Conditional* | Regex patterns of attributes to remove |
| when          | [When](../when.md) | No           | Conditions specifying when to apply    |

*Exactly one of `tag_patterns`, or `attr_patterns` is required.*

## Examples

```yaml
channel_rules:
  - remove_field:
      attr_patterns: ["tvg-logo", "tvg-id"]
      when:
        attr:
          name: "group-title"
          patterns: ["^International$"]
```
