# When Conditions

The `when` clause is a conditional system used to apply rules only to specific channels that match certain criteria. It
provides a powerful filtering mechanism using regex patterns and logical operators to precisely target channels for rule
application.

## YAML Structure

```yaml
when:
  - name: ""
  - tag:
      name: ""
      value: ""
  - attr:
      name: ""
      value: ""
  - and: []
  - or: []
  - not: []
```

## Fields

### Condition Fields

| Field  | Type     | Required | Description                                        |
|--------|----------|----------|----------------------------------------------------|
| `name` | `regex`  | No       | Match against channel name using regex             |
| `tag`  | `object` | No       | Match against M3U tags (e.g., "EXTGRP")            |
| `attr` | `object` | No       | Match against M3U attributes (e.g., "group-title") |

### Logical Operators

| Field | Type          | Required | Description                      |
|-------|---------------|----------|----------------------------------|
| `and` | `[]condition` | No       | All nested conditions must match |
| `or`  | `[]condition` | No       | Any nested condition must match  |
| `not` | `[]condition` | No       | Invert result                    |

### Tag and Attribute Objects

| Field   | Type    | Required | Description                            |
|---------|---------|----------|----------------------------------------|
| `name`  | `regex` | Yes      | Tag or attribute name (regex pattern)  |
| `value` | `regex` | Yes      | Tag or attribute value (regex pattern) |

**Note:** The default behavior for multiple conditions in a `when` array is OR.

## Examples

### Basic Name Matching

```yaml
# Match channels starting with "CNN"
when:
  - name: "^CNN.*"
```

### Attribute Matching

```yaml
# Match channels in "Sports" group
when:
  - attr:
      name: "group-title"
      value: "Sports"
```

### Multiple Conditions (OR - Default)

```yaml
# Any condition can match
when:
  - name: "^MTV.*"
  - name: "^VH1.*"
  - name: "^Music.*"
```

### AND Condition

```yaml
# All conditions must match
when:
  - and:
      - name: "^MTV.*"
      - attr:
          name: "group-title"
          value: "Music"
```

### NOT Condition

```yaml
# Exclude adult content
when:
  - not:
      - attr:
          name: "group-title"
          value: "(?i)adult"
```

### Complex Nested Conditions

```yaml
# Sports channels that are either HD or 4K, but not test channels
when:
  - and:
      - attr:
          name: "group-title"
          value: "Sports"
      - or:
          - name: ".*HD.*"
          - name: ".*4K.*"
      - not:
          - name: "(?i)test.*"
```

### Advanced Regex Examples

```yaml
# Exclude specific patterns
when:
  - not:
      - name: ".*[Tt]est.*"
      - name: ".*[Ss]ample.*"
      - attr:
          name: "group-title"
          value: "(?i)(adult|xxx|18\\+)"
```

### Tag-Based Conditions

```yaml
# Match based on specific M3U tags
when:
  - tag:
      name: "EXTGRP"
      value: "Entertainment"
```