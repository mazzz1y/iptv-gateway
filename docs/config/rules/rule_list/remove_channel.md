# Remove Channel

The `remove_channel` rule completely removes channels from the playlist based on specified conditions. This rule is particularly useful for content filtering, removing test channels, or excluding unwanted content categories.

## YAML Structure

```yaml
remove_channel: {}
```

## Fields

This rule does not accept any configuration fields. It is designed to be used exclusively with the `when` clause to specify which channels should be removed.

| Field  | Type | Required | Description                          |
|--------|------|----------|--------------------------------------|
| (none) | -    | -        | This rule has no configurable fields |

## Examples

### Remove Adult Content

```yaml
remove_channel: {}
when:
  - attr:
      name: "group-title"
      value: "(?i)(adult|xxx|18\\+)"
```

### Remove Test Channels

```yaml
remove_channel: {}
when:
  - or:
      - name: "(?i).*test.*"
      - name: "(?i).*sample.*"
      - name: "(?i).*demo.*"
```

### Remove Low Quality Channels

```yaml
remove_channel: {}
when:
  - name: ".*SD.*"
  - not:
      - name: ".*HD.*|.*4K.*|.*UHD.*"
```

### Remove Specific Group

```yaml
remove_channel: {}
when:
  - attr:
      name: "group-title"
      value: "^Shopping$"
```

### Complex Filtering

```yaml
remove_channel: {}
when:
  - or:
      # Remove adult content
      - attr:
          name: "group-title"
          value: "(?i)(adult|xxx|18\\+|porn)"
      # Remove test channels
      - name: "(?i).*(test|sample|demo|fake).*"
      # Remove inactive channels
      - and:
          - attr:
              name: "group-title"
              value: "(?i).*inactive.*"
          - not:
              - attr:
                  name: "status"
                  value: "active"
```

### Remove Low Resolution Channels

```yaml
remove_channel: {}
when:
  - and:
      - name: ".*480p.*|.*360p.*|.*240p.*"
      - not:
          - attr:
              name: "group-title"
              value: "(?i)retro.*|.*classic.*"
```
