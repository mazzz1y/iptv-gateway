# Cache

The cache block configures the caching layer for IPTV streams and metadata. It handles file downloads and caching logic.

## Key Concepts

- When the TTL expires, the application will attempt to renew the cache. If the file is unchanged (based on Expires,
  Last-Modified, and ETag headers), the TTL will be renewed.
- Retention is used to clean up expired cache files that have not been accessed or renewed for the specified time
  period.
- Compression can be enabled to reduce disk usage by gzipping cached files. Slightly increased CPU usage is expected
  when compression is enabled.

## YAML Structure

```yaml
path: ""
ttl: ""
retention: ""
compression: false
http_headers: []
```

## Fields

| Field          | Type                               | Required | Default    | Description                                            |
|:---------------|:-----------------------------------|:---------|:-----------|:-------------------------------------------------------|
| `path`         | `string`                           | No       | `"/cache"` | Directory path where cache files will be stored        |
| `ttl`          | `string`                           | No       | `"24h"`    | Cache expiration time (e.g., "1h", "30m")              |
| `retention`    | `string`                           | No       | `"30d"`    | How long to keep unaccessed files on disk (e.g., "7d") |
| `compression`  | `boolean`                          | No       | `false`    | Enable gzip compression for cached files               |
| `http_headers` | [`[]NameValue`](#namevalue-object) | No       | `[]`       | Extra request headers for outgoing requests            |

### Name/Value Object

| Field   | Type     | Required | Description                          |
|---------|----------|----------|--------------------------------------|
| `name`  | `string` | Yes      | Name identifier for the object       |
| `value` | `string` | Yes      | Value associated with the given name |
