# Rules

Rules allow you to modify, filter, and transform channels or channel lists using a flexible range of operations. Rules are defined globally and applied to all clients, with optional filtering by user or playlist names.

!!! note "Rule Processing"

* All rules are defined at the global level
* Rules can be filtered to specific users or playlists using `when` conditions
* Channel-level rules are processed first, followed by playlist-level operations

## YAML Structure

```yaml
rules:
  - <rule_name>:
    # rule configuration
```