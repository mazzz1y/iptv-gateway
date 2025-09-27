# Configuration

IPTV Gateway can read configuration from a file or from a directory by combining multiple files based on top-level
elements.

By default, it reads configuration from `config.yaml` in the current directory.

```bash
iptv-gateway -config ./config.yaml # from file
iptv-gateway -config ./config      # from directory
```

!!! note "Hint"

    All arrays with a single value can be specified without brackets.

### Root Level Configuration

| Field           | Type                                       | Description                                                       |
|-----------------|--------------------------------------------|-------------------------------------------------------------------|
| `server`        | [Server](./config/server.md)               | Server configuration including listening addresses and public URL |
| `url_generator` | [URL Generator](./config/url_generator.md) | URL generation and encryption configuration                       |
| `log`           | [Log](config/logs.md)                      | Logging configuration                                             |
| `proxy`         | [Proxy](./config/proxy.md)                 | Stream proxy configuration for remuxing with ffmpeg               |
| `cache`         | [Cache](./config/cache.md)                 | Cache configuration for playlists and EPGs                        |
| `playlists`     | [Playlists](./config/playlists.md)         | Array of playlist definitions with sources                        |
| `epgs`          | [EPGs](./config/epgs.md)                   | Array of EPG definitions with sources                             |
| `rules`         | [Rules](./config/rules/index.md)           | Global processing rules for all clients                           |
| `clients`       | [Clients](./config/clients.md)             | Array of IPTV client definitions with individual settings         |