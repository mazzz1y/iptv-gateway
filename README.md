# Majmun

> ⚠️ **Warning**: This project is currently under construction. Some features might be unstable or incomplete.

A minimal, functional IPTV gateway for your TV

This app allows you to distribute playlists and XMLTV files to your TVs and centrally manage them in a flexible manner.

## Usage

https://mazzz1y.github.io/majmun/

```bash
docker run -d -p 8080:8080 \
  -v $PWD/cache:/cache \
  -v $PWD/config.yaml:/config/config.yaml:ro \
  ghcr.io/mazzz1y/majmun:latest
```