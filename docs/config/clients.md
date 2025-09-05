# Clients

The clients block represents a list of IPTV clients. Each client typically corresponds to a single device or user
accessing the IPTV service. Clients are identified by their unique name and authenticated using a secret key.

## Client Links

Each client can access the following endpoints:

- `{public_url}/{client_secret}/playlist.m3u8`
- `{public_url}/{client_secret}/epg.xml`
- `{public_url}/{client_secret}/epg.xml.gz`

## YAML Structure

```yaml
clients:
  - name: client-name
    secret: "your-secret-key"
    preset: preset1
    playlist: ["playlist1", "playlist2"]
    epg: ["epg1", "epg2"]
```

## Fields

| Field      | Type                   | Required | Description                                  |
|------------|------------------------|----------|----------------------------------------------|
| `name`     | `string`               | Yes      | Unique name identifier for this client       |
| `secret`   | `string`               | Yes      | Authentication secret key for the client     |
| `preset`   | `string` or `[]string` | No       | List of preset names to apply to this client |
| `playlist` | `string` or `[]string` | No       | List of playlist names for this client       |
| `epg`      | `string` or `[]string` | No       | List of EPG names for this client            |

## Examples

### Basic Client Configuration

```yaml
clients:
  - name: living-room-tv
    secret: "secret"
    playlist: "sports-playlist"
```

### Client with Multiple Presets

```yaml
clients:
  - name: family-tablet
    secret: "secret"
    preset: ["family-friendly", "hd-quality"]
    playlist: ["basic-playlist", "kids-playlist"]
```