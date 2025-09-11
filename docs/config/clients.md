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
  - name: ""
    secret: ""
    presets: []
    proxy: {}
    channel_rules: []
    playlist_rules: []
    playlists: []
    epgs: []
```

## Fields

    # proxy: { }  # optional
    # channel_rules: []
    # playlist_rules: []

| Field            | Type       | Required | Description                                  |
|------------------|------------|----------|----------------------------------------------|
| `name`           | `string`   | Yes      | Unique name identifier for this client       |
| `secret`         | `string`   | Yes      | Authentication secret key for the client     |
| `presets`        | `[]string` | No       | List of preset names to apply to this client |
| `playlists`      | `[]string` | No       | List of playlist names for this client       |
| `epgs`           | `[]string` | No       | List of EPG names for this client            |
| `proxy`          | `object`   | No       | Optional per-client proxy config             |
| `channel_rules`  | `array`    | No       | Per-client channel rules                     |
| `playlist_rules` | `array`    | No       | Per-client playlist rules                    |

## Examples

### Basic Client Configuration

```yaml
clients:
  - name: living-room-tv
    secret: "living-room-secret-123"
    playlists: "sports-playlist"
```

### Client with Multiple Presets

```yaml
clients:
  - name: family-tablet
    secret: "family-tablet-secret-456"
    presets: ["family-friendly", "hd-quality"]
    playlists: ["basic-playlist", "kids-playlist"]
```
