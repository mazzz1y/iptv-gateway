# Sort Playlist Rule

The `sort` rule controls the ordering of channels in a playlist, optionally grouping by attributes or tags and allowing full regex ordering.

!!! note
    This rule applies to the entire channel list after channel-level rules are processed.

## YAML Structure

```yaml
playlist_rules:
  - sort:
      selector:
        attr: group-title
      order:
        - ".*HD.*"
        - ""
        - ".*UHD.*"
      group_by:
        selector:
          attr: category
        order:
          - "News"
          - "Sports"
          - ""
```

**Simple alphabetical sort:**
```yaml
playlist_rules:
  - sort: {}
```

## Fields

| Field        | Type                           | Required | Description                                                      |
|--------------|--------------------------------|----------|------------------------------------------------------------------|
| selector     | [`Selector`](../common.md)     | No       | Property to use for sorting (attribute/tag/etc), default is name |
| order        | `[]regex`                      | No       | Custom order of channels, regex patterns                         |
| group_by     | [`GroupByRule`](#groupbyrule)  | No       | Group before sorting                                             |

### GroupByRule

| Field      | Type                           | Required | Description                                |
|------------|--------------------------------|----------|--------------------------------------------|
| selector   | [`Selector`](../common.md)     | Yes      | How to group (attribute/tag)               |
| order      | `[]regex`                      | No       | Custom order of groups, regex patterns      |

## How It Works

1. If `group_by` is set, channels are grouped by `selector`.
2. Channel or group order is determined by the corresponding `order` arrays (if present).
3. Within each group (or globally), regex `order` is applied in order. Unmatched go at the end.
4. Channels within the same priority are alphabetically sorted by name/selector value.

## Examples
### Sorted by attribute, within group order
```yaml
playlist_rules:
  - sort:
      selector:
        attr: tvg-name
      order: ["^A", "^B", ""]
      group_by:
        selector:
          tag: group-title
        order:
          - "News"
          - "Children"
          - ""
```
