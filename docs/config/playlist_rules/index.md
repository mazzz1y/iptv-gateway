# Playlist Rules

Playlist rules are processing instructions that operate on the entire playlist structure.

!!! note "Rule Processing Order"

    Global Rules ➡ Subscription Rules ➡ Preset Rules ➡ Client Rules

## YAML Structure

```yaml
playlist_rules:
  - rule_name:
      - # rule configuration
```