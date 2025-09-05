# Log

Log configuration controls the application's logging behavior, including the log level and output format.

## YAML Structure

```yaml title="Default"
log:
  level: "info"
  format: "text"
```

## Configuration Fields

| Field    | Type     | Required | Description                              |
|----------|----------|----------|------------------------------------------|
| `level`  | `string` | No       | Logging level (debug, info, warn, error) |
| `format` | `string` | No       | Log output format (text, json)           |

## Log Levels

| Level   | Description                                             |
|---------|---------------------------------------------------------|
| `debug` | Most verbose level, includes all log messages           |
| `info`  | General information messages (default)                  |
| `warn`  | Warning messages for potentially problematic situations |
| `error` | Error messages for serious problems                     |

## Log Formats

| Format | Description                          |
|--------|--------------------------------------|
| `text` | Human-readable text format (default) |
| `json` | JSON structured logging format       |
