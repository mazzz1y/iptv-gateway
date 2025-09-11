# Channel Rules

Channel rules operate at the channel level and can modify or transform channel properties

!!! note "Rule Processing Order"

    Global Rules ➡ Subscription Rules ➡ Preset Rules ➡ Client Rules

## YAML Structure

```yaml
channel_rules:
  - <rule_name>:
      # rule configuration
      when: {}
```

## Fields

*For other fields, see specific rule types.*

| Field | Type            | Required | Description              |
|-------|-----------------|----------|--------------------------|
| when  | [When](when.md) | No       | Conditions to apply rule |
