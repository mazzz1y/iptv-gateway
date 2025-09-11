# Playlist Rules

Playlist rules operate on the entire playlist and can modify or filter channel lists based on configured criteria.

!!! note "Rule Processing Order"

    Global Rules ➡ Subscription Rules ➡ Preset Rules ➡ Client Rules

## YAML Structure

```yaml
playlist_rules:
  - <rule_name>:
      # rule configuration (see below)
```

## Fields

*For other fields, see specific rule types.*