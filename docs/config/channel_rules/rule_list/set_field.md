# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes.

## YAML Structure

```yaml
channel_rules:
  - set_field:
      # exactly one of the following:
      name: "<new name>"          # Set channel name directly
      attr:
        name: "<attr name>"
        value: "<attr value>"     # Set M3U attribute
      tag:
        name: "<tag name>"
        value: "<tag value>"       # Set M3U tag
      when: {}
```

## Fields

| Field | Type                           | Required     | Description                                 |
|-------|--------------------------------|--------------|---------------------------------------------|
| name  | `string`                       | Conditional* | New channel name                            |
| attr  | [`NameValue`](../../common.md) | Conditional* | Set attribute (must include name and value) |
| tag   | [`NameValue`](../../common.md) | Conditional* | Set tag (must include name and value)       |
| when  | [When](../when.md)             | No           | Conditions specifying when to apply         |

*Exactly one of `name`, `attr`, or `tag` is required.*
