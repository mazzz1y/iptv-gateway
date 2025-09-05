# Set Field

The `set_field` rule allows you to modify channel properties including the channel name, M3U tags, and attributes. It
supports Go template syntax for dynamic value generation based on existing channel data.

## YAML Structure

```yaml
set_field:
  - name: ""
  - tag:
      name: ""
      template: ""
  - attr:
      name: ""
      template: ""
```

## Fields

### Main Fields

| Field  | Type         | Required | Description                                         |
|--------|--------------|----------|-----------------------------------------------------|
| `name` | `gotemplate` | No       | Set the channel name using static value or template |
| `tag`  | `object`     | No       | Set M3U tag values                                  |
| `attr` | `object`     | No       | Set M3U attribute values                            |

### Tag and Attribute Objects

| Field      | Type         | Required | Description                         |
|------------|--------------|----------|-------------------------------------|
| `name`     | `string`     | Yes      | Name of the tag or attribute to set |
| `template` | `gotemplate` | Yes      | Value or Go template for the field  |

### Available Template Variables

| Variable        | Type                | Description                   |
|-----------------|---------------------|-------------------------------|
| `Channel.Name`  | `string`            | Current channel name          |
| `Channel.Attrs` | `map[string]string` | Map of all channel attributes |
| `Channel.Tags`  | `map[string]string` | Map of all channel tags       |

## Examples

### Setting Static Channel Name

```yaml
set_field:
  - name: "CNN International"
when:
  - name: "^CNN.*"
```

### Dynamic Channel Name with Template

```yaml
set_field:
  - name: "{{ .Channel.Name }} [HD]"
when:
  - name: "^ESPN.*"
  - not:
      - name: ".*HD.*"
```

### Complex Name Template

```yaml
set_field:
  - name: "{{ .Channel.Attrs.country | default \"Unknown\" }} - {{ .Channel.Name }}"
when:
  - attr:
      name: "group-title"
      value: "International"
```

### Setting Group Title Attribute

```yaml
set_field:
  - attr:
      name: "group-title"
      template: "Sports & Entertainment"
when:
  - name: "^(ESPN|Fox Sports|Sky Sports).*"
```

### Setting Custom Tag

```yaml
set_field:
  - tag:
      name: "EXTGRP"
      template: "{{ .Channel.Attrs.group-title }}"
when:
  - not:
      - tag:
          name: "EXTGRP"
          value: ".*"
```

### Country-Based Grouping

```yaml
set_field:
  - attr:
      name: "group-title"
      template: "{{ .Channel.Attrs.country | default \"International\" }} Channels"
  - name: "[{{ .Channel.Attrs.country | default \"INT\" }}] {{ .Channel.Name }}"
when:
  - attr:
      name: "country"
      value: ".*"
```
