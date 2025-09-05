# Conditions

Conditions are reusable named condition definitions that can be referenced in channel rules and playlist rules.

## YAML Structure

```yaml
conditions:
  - name: condition-name
    when: []
```

## Fields

| Field  | Type          | Required | Description                                                                            |
|--------|---------------|----------|----------------------------------------------------------------------------------------|
| `name` | `string`      | Yes      | Unique name identifier for this condition                                              |
| `when` | `[]condition` | Yes      | Array of condition objects - see [When Conditions](./channel_rules/when.md) for syntax |

## Usage

Named conditions can be referenced in rules using:

- **String reference**: `when: condition_name`
- **Array reference**: `when: ["condition1", "condition2"]`

## Examples

### Basic Condition Definition

```yaml
conditions:
  - name: "adult"
    when:
      - attr:
          name: "group-title"
          value: "(?i)adult"
```

### Using Conditions in Rules

```yaml
# Define conditions
conditions:
  - name: "adult"
    when:
      - attr:
          name: "group-title"
          value: "(?i)adult"

# Use conditions in channel rules
channel_rules:
  # Single condition reference
  - remove_channel: {}
    when: adult

  # Multiple condition references (OR logic)
  - remove_channel: {}
    when: ["adult", "test_channels"]
```
