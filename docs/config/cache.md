# Cache

The cache block configures the caching system for IPTV streams and metadata. Caching improves performance by storing
frequently accessed data locally, reducing load on upstream sources and providing faster response times to clients.

## Key Concepts

- When the TTL expires, the application will attempt to renew the cache. If the file is unchanged (based on Expires,
  Last-Modified, and ETag headers), the TTL will be renewed.
- Retention is used to clean up expired cache files that have not been accessed or renewed for the specified time
  period.

## YAML Structure

```yaml title="Default"
cache:
  path: "/cache"
  ttl: "24h"
  retention: "30d"
```

## Fields

| Field       | Type     | Required | Description                                             |
|-------------|----------|----------|---------------------------------------------------------|
| `path`      | `string` | No       | Directory path where cache files will be stored         |
| `ttl`       | `string` | No       | Cache expiration time (e.g., "1h", "30m")               |
| `retention` | `string` | No       | How long to keep unaccessed files on disk (e.g., "7d"). |



