### Final Value

Final value allows customizing the result channel after removing duplicates. Used in deduplication rules.

```yaml
final_value:
  selector: {}
  template: ""
```

#### Template values

| Variable             | Description            |
|----------------------|------------------------|
| `{{.Channel.Name}}`  | Original channel name  |
| `{{.Channel.Attrs}}` | Channel attributes map |
| `{{.Channel.Tags}}`  | Channel tags map       |
| `{{.BaseName}}`      | Duplicates basename    |