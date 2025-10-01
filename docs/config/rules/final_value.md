# Final Value

Final value allows customizing the result channel after merging/removing duplicates. Used in deduplication rules.

```yaml
final_value:
  selector: {}
  template: ""
```

## Template Variables

| Variable                  | Type              | Description                                               |
|---------------------------|-------------------|-----------------------------------------------------------|
| `{{.Channel.Name}}`       | string            | The original channel name.                                |
| `{{.Channel.Attrs}}`      | map[string]string | A map containing the channel's attributes.                |
| `{{.Channel.Tags}}`       | map[string]string | A map containing the channel's tags.                      |
| `{{.Channel.BaseName}}`   | string            | Duplicates basename                                       |
| `{{.Playlist.Name}}`      | string            | The best channel's playlist name.                         |
| `{{.Playlist.IsProxied}}` | bool              | Indicates whether the best channel's playlist is proxied. |