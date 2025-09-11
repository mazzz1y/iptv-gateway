# Server

The server block is responsible for configuring HTTP servers.

## YAML Structure

```yaml title="Default"
server:
  listen_addr: ":6078"
  metrics_addr: ""
  public_url: "http://127.0.0.1:6078"
```

## Configuration Fields

| Field          | Type     | Required | Description                                           |
|----------------|----------|----------|-------------------------------------------------------|
| `listen_addr`  | `string` | Yes      | Address the gateway listens on                        |
| `public_addr`  | `string` | Yes      | Public address of the gateway, used to generate links |
| `metrics_addr` | `string` | No       | Address for the metrics server, disabled if empty     |
