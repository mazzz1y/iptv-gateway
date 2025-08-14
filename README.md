# IPTV Gateway

> ⚠️ **Warning**: This project is currently under construction. Some features might be unstable or incomplete.

A minimal, functional IPTV gateway for your TV

This app allows you to distribute playlists and XMLTV files to your TVs and centrally manage them in a flexible manner.

## Features

- Proxying M3U8 and EPG, including stream remuxing
- Playlist merging
- Filtering based on M3U8 tags, attributes, or channel names (global, per subscription, or per client)
- On-disk caching
- Custom error screens (specify any streaming command)
- Rate limiting (global, per subscription, or per client)
- Low memory usage — all processing is done on the fly

## Usage

```bash
docker run -d -p 8080:8080 \
  -v $PWD/cache:/cache \
  -v $PWD/config.yaml:/config/config.yaml:ro \
  ghcr.io/mazzz1y/iptv-gateway:latest
```

```yaml
# Proxy, Rules, and Concurrent Streams can be set at the global, subscription, or user level.
# Users can use the following links:
# - http://localhost:8080/{secret}/playlist.m3u8
# - http://localhost:8080/{secret}/epg.xml.gz
# - http://localhost:8080/{secret}/epg.xml


listen_addr: ":8080"
public_url: "http://localhost:8080"

secret: secret # Secret used for encrypting proxy links

proxy:
  # Enable ffmpeg remuxing globally
  enabled: true
  concurrency: 4


presets:
  - name: children
    rules:
      - remove_channel: { }
        when:
          - attr:
              name: "tvg-group"
              value: "adult"

subscriptions:
  - name: english
    playlist: http://example.com/english.m3u8 # Both playlist and epg can be arrays of links; they will be merged
    epg: http://example.com/en.xml.gz
    proxy:
      concurrency: 2

  - name: french
    playlist: https://example.com/french.m3u
    epg: http://example.com/fr.xml.gz
    proxy:
      concurrency: 3

clients:
  - name: device1
    secret: secret1
    subscriptions: french
    rules:
      - remove_channel: {}
        when:
          - attr:
              name: "tvg-language"
              value: "^english$"
    channel_name: "Some-Channel"
  - name: device2
    secret: secret2
    preset: children
    subscriptions: [ "english", "french" ] # Both subscriptions will be merged
```

### Custom streaming settings

```yaml
proxy:
  stream:
    command: [ "ffmpeg", "-i", "{{.url}}", "-c", "copy", "-f", "mpegts", "pipe:1" ]
    env_vars:
      http_proxy: http://192.168.1.1:1080
  error:
    command: [ "...", "{{.message}}", "..." ]
    rate_limit_exceeded:
      template_vars:
        message: "Rate limit exceeded. Please try again later."
    link_expired:
      template_vars:
        message: "Link has expired. Please refresh your playlist."
    upstream_error:
      template_vars:
        message: "Unable to play stream. Please try again later or contact administrator."
```

### Rules format

```yaml
rules:
  - remove_channel: {}
    when:
      - "name": ".*Test.*"

  - remove_field:
      - type: "attr"
        name: "tvg-logo"

  - set_field:
      - type: "attr"
        name: "original-name"
        template: "{{ .Channel.Name }}"

      - type: "name"
        template: "{{ .Channel.Name | title }}"

      - type: "tag"
        name: "EXTGRP"
        template: "{{ index .Channel.Attrs \"tvg_group\" | upper }}"
    when:
      - attr:
          name: "tvg-group"
          value: [ "entertainment" ]

  - set_field:
      - type: "attr"
        name: "display-name"
        template: |
          {{- if index .Channel.Tags "EXTGRP" | eq "4K" -}}
            🎬 {{ .Channel.Name | title }}
          {{- else -}}
            📺 {{ .Channel.Name | title }}
          {{- end -}}

```