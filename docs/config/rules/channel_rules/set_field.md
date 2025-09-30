# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes.

## YAML Structure

```yaml
set_field:
  selector: {}
  template: ""
  condition: {}
```

## Fields

| Field       | Type                           | Required | Description                               |
|-------------|--------------------------------|----------|-------------------------------------------|
| `selector`  | [`Selector`](../selector.md)   | Yes      | What property to set (attribute/tag/name) |
| `template`  | `gotemplate`                   | Yes      | The template definition for the new value |
| `condition` | [`Condition`](../condition.md) | No       | Optional, restricts rule activation       |

#### Template values

| Variable             | Description            |
|----------------------|------------------------|
| `{{.Channel.Name}}`  | Original channel name  |
| `{{.Channel.Attrs}}` | Channel attributes map |
| `{{.Channel.Tags}}`  | Channel tags map       |

## Examples

Set channel group to `Free` for all channels in the `custom` playlist:

```yaml
channel_rules:
  - set_field:
      selector: tag/EXTGRP
      template: "Free"
      condition:
        playlists: custom
```

