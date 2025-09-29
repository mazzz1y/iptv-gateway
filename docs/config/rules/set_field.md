# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes.

## YAML Structure

```yaml
channel_rules:
  - set_field:
      selector: {...} # channel property selector, see below
      template: {...} # template definition for new value, see below
      condition: {...} # optional, see [Condition](condition.md) section

```

## Fields

| Field         | Type                         | Required  | Description                               |
|-------------- |-----------------------------|-----------|-------------------------------------------|
| `selector`    | [`Selector`](../common.md)   | Yes       | What property to set (attribute/tag/name) |
| `template`    | [`Template`](../common.md)   | Yes       | The template definition for the new value |
| `condition`   | [`Condition`](./condition.md) | No        | Optional, restricts rule activation       |


## Template

List of available template variables:

- `{{.Channel.Name}}` - Original channel name
- `{{.Channel.Attrs}}` - Channel attributes map
- `{{.Channel.Tags}}` - Channel tags map
