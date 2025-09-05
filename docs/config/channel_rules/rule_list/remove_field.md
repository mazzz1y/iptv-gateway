# Remove Field

The `remove_field` rule allows you to delete specific fields from channels, including the channel name, M3U tags, and
attributes. This is useful for cleaning up unwanted metadata or removing problematic fields.

## YAML Structure

```yaml
remove_field:
  - name: ""
  - tag: []
  - attr: []
```

## Fields

| Field  | Type                 | Required | Description                                           |
|--------|----------------------|----------|-------------------------------------------------------|
| `name` | `any`                | No       | Remove the channel name (set to any value to remove)  |
| `tag`  | `regex` or `[]regex` | No       | Array of regex patterns for tag names to remove       |
| `attr` | `regex` or `[]regex` | No       | Array of regex patterns for attribute names to remove |

## Examples

### Remove Channel Name

```yaml
remove_field:
  - name: true
when:
  - name: "^Test.*"
```

### Remove Specific Attribute

```yaml
remove_field:
  - attr: "group-title"
when:
  - name: "^Sample.*"
```

### Remove Multiple Attributes

```yaml
remove_field:
  - attr: ["tvg-logo", "tvg-id", "tvg-chno"]
when:
  - attr:
      name: "group-title"
      value: "(?i)test.*"
```

### Remove All Logo Attributes

```yaml
remove_field:
  - attr: ".*logo.*"
```

### Remove Country-Specific Attributes

```yaml
remove_field:
  - attr: ["country", "language", "region"]
when:
  - attr:
      name: "group-title"
      value: "International"
```
