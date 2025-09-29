# Remove Duplicates

The `remove_duplicates` rule keeps only one channel among groups considered duplicates (e.g., preferring "CNN HD" over "CNN").

## YAML Structure

```yaml
playlist_rules:
  - remove_duplicates:
      patterns: ["4K", "UHD", "FHD", "HD", "SD"] # in priority order
      selector:
        attr: "tvg-id"  # or tag, or leave empty to use name
      set_field: {...}   # (optional) see set_field docs
      condition: {...}   # (optional)
```

## Fields

| Field      | Type                         | Required | Description                                                          |
|------------|------------------------------|----------|----------------------------------------------------------------------|
| patterns   | `[]regex`                    | Yes      | Priority order (first has highest priority)                          |
| selector   | [`Selector`](../common.md)   | No       | Use attribute or tag to identify groups                              |
| set_field  | [`SetField`](set_field.md)   | No       | Template/object for output name/attrs                                |
| condition  | [`Condition`](condition.md)       | No       | Only apply if condition matches                                      |


## Examples

Prefer best quality:

```yaml
playlist_rules:
  - remove_duplicates:
      patterns: ["4K", "UHD", "FHD", "HD", "SD"]
```

Restrict deduplication to specific clients:

```yaml
playlist_rules:
  - remove_duplicates:
      patterns: ["SD", "HD", "FHD", "4K"]
      condition:
        clients: ["kitchen", "office-lite"]
```
