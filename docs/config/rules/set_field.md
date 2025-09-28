# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes.

## YAML Structure

```yaml
channel_rules:
  - set_field:
      set_field:
        # exactly one of the following:
        name_template: "template"        # Set channel name using template
        attr_template:
          name: "<attr name>"
          template: "template"           # Set M3U attribute using template
        tag_template:
          name: "<tag name>"
          template: "template"           # Set M3U tag using template
      when: {}
```

## Fields

| Field                      | Type                           | Required     | Description                         |
|----------------------------|--------------------------------|--------------|-------------------------------------|
| `set_field.name_template`  | `template`                     | Conditional* | Template for channel name           |
| `set_field.attr_template`  | [`NameTemplate`](../common.md) | Conditional* | Set attribute using template        |
| `set_field.tag_template`   | [`NameTemplate`](../common.md) | Conditional* | Set tag using template              |
| `when`                     | [When](when.md)                | No           | Conditions specifying when to apply |

*Exactly one of `name_template`, `attr_template`, or `tag_template` is required.*

## Template

List of available template variables:

- `{{.Channel.Name}}` - Original channel name
- `{{.Channel.Attrs}}` - Channel attributes map
- `{{.Channel.Tags}}` - Channel tags map
