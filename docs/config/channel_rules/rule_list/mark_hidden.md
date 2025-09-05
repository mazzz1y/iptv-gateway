# Mark Hidden

The `mark_hidden` rule marks channels as hidden, removing them from metrics and logs while keeping them in the playlist.

## YAML Structure

```yaml
mark_hidden: true
```

## Fields

This rule does not accept any configuration fields. It is designed to be used exclusively with the `when` clause to
specify which channels should be marked as hidden.

| Field  | Type | Required | Description                          |
|--------|------|----------|--------------------------------------|
| (none) | -    | -        | This rule has no configurable fields |

## Examples

### Hide Test Channels from Monitoring

```yaml
mark_hidden: true
when:
  - name: "(?i).*test.*"
  - name: "(?i).*sample.*"
  - name: "(?i).*demo.*"
```