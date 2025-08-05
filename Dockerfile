FROM golang:1.24-alpine3.22 AS build

ENV CGO_ENABLED=0
COPY . /src

RUN cd /src && \
  go build -ldflags="-s -w" -trimpath -o /iptv-gateway ./cmd/iptv-gateway

FROM alpine:3.22

RUN apk add --no-cache ffmpeg

COPY --from=build /iptv-gateway /iptv-gateway
USER 1337

ENTRYPOINT ["/iptv-gateway"]
