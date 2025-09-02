# Proxy

The proxy block configures the streaming proxy functionality, also known as "remuxing".
This feature allows the gateway to act as an intermediary between IPTV clients and upstream sources, providing stream processing, transcoding, and error handling capabilities.

When proxying is enabled, the links in the playlist will be encrypted and will point to the IPTV Gateway app.

The default configuration uses FFmpeg for remuxing and is ready to use out of the box. Most users can enable proxy functionality by simply setting `enabled` to `true`. Advanced users can customize commands to add transcoding, filtering, or other stream processing features.
!!! note "Rule Merging Order"

    Proxy can be defined at multiple levels in the configuration. It will be merged in the following order, with each level overriding the previous one:

    Global Proxy ➡ Subscription Proxy ➡ Preset Proxy ➡ Client Proxy

    This applies to all proxy-related fields, **except concurrency**.

!!! note "Concurrency Handling"

    Concurrency is handled at the global, subscription, and client levels separately.

## YAML Structure

```yaml
proxy:
  enabled: false
  concurrency: 0
  stream:
    command: []
    template_vars: {}
    env_vars: {}
  error:
    command: []
    template_vars: {}
    env_vars: {}
    upstream_error:
      command: []
      template_vars: {}
      env_vars: {}
    rate_limit_exceeded:
      command: []
      template_vars: {}
      env_vars: {}
    link_expired:
      command: []
      template_vars: {}
      env_vars: {}
```

## Fields

### Main Proxy Configuration

| Field         | Type      | Required | Description                                        |
|---------------|-----------|----------|----------------------------------------------------|
| `enabled`     | `bool`    | No       | Enable or disable proxy functionality             |
| `concurrency` | `int`     | No       | Maximum concurrent streams (0 = unlimited)        |
| `stream`      | `command` | No       | Command configuration for stream processing       |
| `error`       | `command` | No       | Default error handling configuration              |

### Command Object

| Field           | Type                | Required | Description                              |
|-----------------|---------------------|----------|------------------------------------------|
| `command`       | `[]gotemplate`      | No       | Command array to execute                 |
| `template_vars` | `map[string]any`    | No       | Variables available in command templates |
| `env_vars`      | `map[string]string` | No     | Environment variables for the command    |

### Error Handling Objects

| Field                   | Type      | Required | Description                                    |
|-------------------------|-----------|----------|------------------------------------------------|
| `upstream_error`        | `command` | No       | Command to run when upstream source fails     |
| `rate_limit_exceeded`   | `command` | No       | Command to run when rate limits are hit       |
| `link_expired`          | `command` | No       | Command to run when stream links expire       |

### Available Template Variables

| Variable        | Type                | Description |
|-----------------|---------------------|-------------|
| `url`           | `string`            | Stream URL  |


## Examples

### Basic Proxy Setup

```yaml
proxy:
  enabled: true
  concurrency: 10
```

### Custom FFmpeg Configuration

```yaml
proxy:
  enabled: true
  concurrency: 5
  stream:
    command:
      - "ffmpeg"
      - "-i"
      - "{{ .url }}"
      - "-c:v"
      - "libx264"
      - "-preset"
      - "ultrafast"
      - "-f"
      - "mpegts"
      - "pipe:1"
    env_vars:
      FFMPEG_LOG_LEVEL: "error"
```

### Error Handling with Test Pattern

```yaml
proxy:
  enabled: true
  error:
    upstream_error:
      command:
        - "ffmpeg"
        - "-f"
        - "lavfi"
        - "-i"
        - "testsrc2=size=1280x720:rate=25"
        - "-f"
        - "lavfi"
        - "-i"
        - "sine=frequency=1000:duration=0"
        - "-c:v"
        - "libx264"
        - "-c:a"
        - "aac"
        - "-t"
        - "3600"
        - "-f"
        - "mpegts"
        - "pipe:1"
```