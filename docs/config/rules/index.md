# Rules

Rules allow you to modify, filter, and transform channels or channel lists using a flexible range of operations.

!!! note "Rule Processing Order"

* Playlist Rules ➡ Preset Rules ➡ Client Rules ➡ Global Rules
* Rules that apply to the entire channel list are processed after the others

## YAML Structure

```yaml
rules:
  - <rule_name>:
    # rule configuration
```