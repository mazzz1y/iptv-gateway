# Configuration Examples

This page provides example configurations to help users get started quickly.

## Simple Configuration

A minimal setup with a basic server, one playlist, and one EPG.

```yaml
server:
  listen_addr: ":8080"
  public_url: "http://localhost:8080"

proxy:
  enabled: true # Proxy everything throw the gateway

playlists:
  - name: basic-tv
    sources: "https://provider.com/basic.m3u8"

epgs:
  - name: tv-guide
    sources: "https://provider.com/guide.xml"

clients:
  - name: "tv"
    secret: "tv-secret"
    playlists: "basic-tv"
    epgs: "tv-guide"
```

## Advanced Configuration

A full-featured setup including proxying, presets, channel rules, playlist rules, and multiple sources.

```yaml
log:
  level: debug
  format: json

server:
  listen_addr: ":8080"
  metrics_addr: ":9090"
  public_url: "https://iptv.example.com"

url_generator:
  secret: "super-secret"
  stream_ttl: "24h"
  file_ttl: "0s"

cache:
  path: "/var/cache/iptv"
  ttl: "12h"
  retention: "7d"
  compression: true

proxy:
  enabled: true
  concurrency: 10 # Set global concurrency
  error: # Override stream error templates
    upstream_error:
      template_vars:
        - name: message
          value: |
            Canal temporalmente no disponible

            No hay respuesta del servidor ascendente

            Por favor, inténtelo más tarde
    rate_limit_exceeded:
      template_vars:
        - name: message
          value: |
            Se ha excedido el número de transmisiones simultáneas

            Por favor, inténtelo más tarde
    link_expired:
      template_vars:
        - name: message
          value: |
            El enlace del canal ha expirado

            Por favor, actualice la lista de reproducción en su televisor

playlists:
  - name: movies
    sources:
      - "https://provider.com/movies1.m3u8"
      - "https://provider.com/movies2.m3u8"
  - name: tv
    sources:
      - "https://provider.com/tv1.m3u8"
      - "https://provider.com/tv2.m3u8"
  - name: sports
    sources:
      - "https://provider.com/sports1.m3u8"
      - "https://provider.com/sports2.m3u8"

epgs:
  - name: movies
    sources:
      - "https://movies.com/guide.xml"
      - "https://movies2.com/guide.xml.gz"
  - name: tv
    sources:
      - "https://tv.com/guide.xml"
      - "https://tv2.com/guide.xml.gz"
  - name: sports
    sources:
      - "https://sports.com/guide.xml"
      - "https://sports2.com/guide.xml.gz"

presets:
  - name: sd-entertainment
    playlists: ["tv", "movies"]
    epgs: ["tv", "movies"]
    playlist_rules:
      - remove_duplicates:
          name: ["SD", "HD", "FHD", "4K"] # Prefer SD quality

  - name: hd-entertainment
    playlists: ["tv", "movies"]
    epgs: ["tv", "movies"]
    playlist_rules:
      - remove_duplicates:
          name: ["4K", "FHD", "HD", "SD"] # Prefer highest quality available

  - name: sports-hd
    playlists: ["sports", "tv"]
    epgs: ["sports", "tv"]

clients:
  - name: "living-room"
    secret: "lr-secret"
    presets: "hd-entertainment"

  - name: "bedroom"
    secret: "br-secret"
    presets: "sports-hd"

  - name: "mobile"
    secret: "mb-secret"
    presets: "sd-entertainment"

  - name: "kitchen"
    secret: "kt-secret"
    presets: "sd-entertainment"
```
