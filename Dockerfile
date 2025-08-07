FROM golang:1.24 AS build

ENV CGO_ENABLED=0
COPY . /src

RUN cd /src && \
  go build -ldflags="-s -w" -trimpath -o /iptv-gateway ./cmd/iptv-gateway

FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /iptv-gateway /iptv-gateway

USER 1337

ENTRYPOINT ["/iptv-gateway"]