# Rules

Rules are the core processing engine of the IPTV Gateway, allowing you to transform, filter, and manipulate IPTV playlists according to your specific requirements. Each rule operates on channels within a playlist and can modify channel properties, remove unwanted content, or perform advanced filtering operations.


!!! note "Rule Processing Order"

    Global Rules ➡ Subscription Rules ➡ Preset Rules ➡ Client Rules


## Key Concepts

- **Sequential Processing**: Rules execute in the order they are defined in the configuration
- **Array-Based**: Each rule accepts an array of configurations, effectively allowing multiple instances of the same rule
- **Conditional Execution**: Rules can be combined with `when` clauses for precise targeting
- **Regex Support**: Most string fields support regular expressions for flexible pattern matching

## YAML Structure

```yaml
rules:
  - rule_name:
      - # rule configuration
    when:
      - # condition configuration
  - another_rule:
      - # another rule configuration
```

## Fields

### Rules Array

| Field   | Type     | Required | Description                              |
|---------|----------|----------|------------------------------------------|
| `rules` | `[]rule` | No       | Array of rule objects to apply in order |

### Single Rule Structure

| Field         | Type   | Required | Description                                    |
|---------------|--------|----------|------------------------------------------------|
| `<rule_name>` | `rule` | Yes      | The specific rule configuration (see individual rule docs) |
| `when`        | `when` | No       | Optional conditions for when to apply this rule |
