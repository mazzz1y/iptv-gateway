# Clients

The clients block represents a list of IPTV clients. Each client typically corresponds to a single device or user accessing the IPTV service. Clients are identified by their unique name and authenticated using a secret key.


## Client Links

Each client can access the following endpoints:

- `{public_url}/{client_secret}/playlist.m3u8`
- `{public_url}/{client_secret}/epg.xml`
- `{public_url}/{client_secret}/epg.xml.gz`

## YAML Structure

```yaml
clients:
  client_name:
    secret: "your-secret-key"
    presets: ["preset1", "preset2"]
    subscriptions: ["sub1", "sub2"]
```

## Fields

| Field           | Type       | Required | Description                                    |
|-----------------|------------|----------|------------------------------------------------|
| `secret`        | `string`   | Yes      | Authentication secret key for the client      |
| `presets`       | `[]string` | No       | List of preset names to apply to this client  |
| `subscriptions` | `[]string` | No       | List of subscription names for this client    |

## Examples

### Basic Client Configuration

```yaml
clients:
  living_room_tv:
    secret: "secret"
    subscriptions: ["sports-package"]
```

### Client with Multiple Presets

```yaml
clients:
  family_tablet:
    secret: "secret"
    presets: ["family-friendly", "hd-quality"]
    subscriptions: ["basic-package", "kids-channels"]
```

### Multiple Clients

```yaml
clients:
  bedroom_tv:
    secret: "secret1"
    presets: ["adult-filter"]
    subscriptions: ["premium-sports"]
  
  kids_tablet:
    secret: "secret2"
    presets: ["child-safe"]
    subscriptions: ["cartoon-channels"]
```