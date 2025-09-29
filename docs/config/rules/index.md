# Rules

Rules allow you to modify, filter, and transform channels or channel lists using a flexible range of operations. Rules
are defined globally and applied to all clients, with optional filtering by client or playlist names.

## Rule Types

Rules are organized into two categories:

- **Channel Rules** - Operate on individual channels (set_field, remove_field, remove_channel, mark_hidden)
- **Playlist Rules** - Operate on the entire playlist/channel list (remove_duplicates, merge_channels, sort)

!!! note "Rule Processing"

* All rules are defined at the global level under `channel_rules` and `playlist_rules`
* Rules can be filtered to specific channels, clients, or playlists using `condition` blocks.
* Channel rules are processed first, followed by playlist rules

## YAML Structure

```yaml
global:
  channel_rules:
    - set_field:
        selector: ...
        template: ...
        condition: {...}
    - remove_field:
        selector: ...
        patterns: [...]
        condition: {...}
    - remove_channel:
        condition: {...}
    - mark_hidden:
        condition: {...}

  playlist_rules:
    - remove_duplicates:
        condition: {...}
    - ...
```

For details on rule objects and the condition system, see below.


## Condition Blocks

Rules may include a `condition` block to restrict when the rule applies. See [Condition Blocks](condition.md).

The `condition` block supports:
- Matching by channel name, client, playlist
- Regex pattern matching
- Attribute/tag selectors
- Logical AND/OR nesting and inversion (`invert`)

All fields are defined as arrays unless otherwise specified. See the [Condition documentation](condition.md) for syntax and options.
