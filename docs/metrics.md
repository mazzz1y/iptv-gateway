# Metrics

The IPTV Gateway provides comprehensive Prometheus metrics for monitoring system performance, stream statistics, and
client activity.

## YAML Configuration

```yaml
metrics_addr: ""
```

## Configuration Fields

| Field          | Type     | Required | Default | Description                         |
|:---------------|:---------|:---------|:--------|:------------------------------------|
| `metrics_addr` | `string` | No       | `""`    | Address and port for metrics server |

## Available Metrics

### Stream Metrics

| Metric Name                        | Type    | Description                           | Labels                                                     |
|------------------------------------|---------|---------------------------------------|------------------------------------------------------------|
| `iptv_subscription_streams_active` | Gauge   | Currently active subscription streams | `subscription_name`                                        |
| `iptv_client_streams_active`       | Gauge   | Currently active client streams       | `client_name`, `subscription_name`, `channel_id`           |
| `iptv_streams_reused_total`        | Counter | Total number of reused streams        | `subscription_name`, `channel_id`                          |
| `iptv_streams_failures_total`      | Counter | Total number of stream failures       | `client_name`, `subscription_name`, `channel_id`, `reason` |

### Request Metrics

| Metric Name                    | Type    | Description                                | Labels                                        |
|--------------------------------|---------|--------------------------------------------|-----------------------------------------------|
| `iptv_listing_downloads_total` | Counter | Total listing downloads by client and type | `client_name`, `listing_type`                 |
| `iptv_proxy_requests_total`    | Counter | Total proxy requests by client and status  | `client_name`, `request_type`, `cache_status` |

## Common Label Values

| Label               | Description                                     | Possible Values                           |
|---------------------|-------------------------------------------------|-------------------------------------------|
| `client_name`       | Unique identifier for each client configuration | any                                       |
| `subscription_name` | Name of the subscription being accessed         | any                                       |
| `channel_id`        | Unique identifier for individual channels       | any                                       |
| `listing_type`      | Type of listing                                 | `playlist`, `epg`                         |
| `request_type`      | Type of request                                 | `file`, `playlist`, `epg`                 |
| `cache_status`      | Cache hit status                                | `hit`, `miss`, `renewed`                  |
| `reason`            | Failure reason                                  | `timeout`, `upstream_error`, `rate_limit` |
