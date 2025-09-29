# Merge Duplicates

The `merge_duplicates` rule combines channels considered duplicates (e.g., "CNN HD", "CNN 4K") into one logical channel with fallback sources.

## YAML Structure

```yaml
playlist_rules:
  - merge_duplicates:
      patterns: ["4K", "HD", "UHD", "FHD"] # regex patterns for names
      selector:
        attr: "tvg-id" # (optional) match attribute, e.g., tvg-id
      set_field: {...} # (optional) how to format output, see set_field docs
      condition: {...} # (optional) see [Condition](../rules/when.md)
```

## Fields

| Field      | Type                         | Required | Description                                                        |
|------------|------------------------------|----------|--------------------------------------------------------------------|
| patterns   | `[]regex`                    | Yes      | List of regex for names/attributes to group                        |
| selector   | [`Selector`](../common.md)   | No       | Match channels using this selector (by default, uses name only)    |
| set_field  | [`SetField`](set_field.md)   | No       | Template/object for output name/attrs                              |
| condition  | [`Condition`](condition.md)       | No       | Only apply to playlists/clients/attributes matching this condition |


## Examples

```yaml
playlist_rules:
  - merge_duplicates:
      patterns: ["4K", "UHD", "FHD", "HD"]
      set_field:
        selector:
          name: true
        template:
          template: "{{.BaseName}} Multi-Quality"
  - merge_duplicates:
      patterns: ["SD", "HD", "FHD"]
      condition:
        clients: ["test", "other-client"]
```
