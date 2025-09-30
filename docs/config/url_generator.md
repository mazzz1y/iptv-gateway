# URL Generator

The URL generator configuration manages the creation and encryption of streaming URLs.

## YAML Structure

```yaml
url_generator:
  secret: ""
  stream_ttl: ""
  file_ttl: ""
```

## Fields

| Field        | Type       | Required | Default             | Description                                         |
|--------------|------------|----------|---------------------|-----------------------------------------------------|
| `secret`     | `string`   | Yes      | ``                  | Secret salt used for URL encryption                 |
| `stream_ttl` | `duration` | No       | `7 days`            | Time-to-live for streaming URLs (0 = no expiration) |
| `file_ttl`   | `duration` | No       | `0 (no expiration)` | Time-to-live for file URLs (0 = no expiration)      |

!!! note "Secret Key"
    This is a salt added to the user's secrets. Changing it will invalidate all links.

!!! note "TTL"
    Setting TTL > 0 will cause links to regenerate each time they're accessed. By default, it's 0, since it's usually
    unnecessary for non-sensitive files.

## Duration Format

Duration values support the following units:

- `s` - seconds
- `m` - minutes
- `h` - hours
- `d` - days (24 hours)

Examples: `30s`, `5m`, `2h`, `1d`, `24h30m`

## URL Generation

The URL generator creates encrypted URLs in the following format:

```
{public_url}/f{extension}
```

Where:

- `{encrypted_token}` contains encrypted stream information and expiration time
- `{extension}` is determined by the content type (`.ts` for streams, original extension for files)