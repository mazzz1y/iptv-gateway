# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes.

## YAML Structure

```yaml
channel_rules:
  - set_field:
      # exactly one of the following:
      name_template: "template"        # Set channel name using template
      attr:
        name: "<attr name>"
        template: "template"           # Set M3U attribute using template
      tag:
        name: "<tag name>"
        template: "template"           # Set M3U tag using template
      when: {}
```

## Fields

| Field           | Type                           | Required     | Description                         |
|-----------------|--------------------------------|--------------|-------------------------------------|
| `name_template` | `template`                     | Conditional* | Template for channel name           |
| `attr`          | [`NameTemplate`](../common.md) | Conditional* | Set attribute using template        |
| `tag`           | [`NameTemplate`](../common.md) | Conditional* | Set tag using template              |
| `when`          | [When](when.md)                | No           | Conditions specifying when to apply |

*Exactly one of `name_template`, `attr`, or `tag` is required.*

## Template

List of available template variables:

- `{{.Channel.Name}}` - Original channel name
- `{{.Channel.Attrs}}` - Channel attributes map
- `{{.Channel.Tags}}` - Channel tags map
