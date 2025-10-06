FROM golang:1.25 AS build

ENV CGO_ENABLED=0
COPY . /src

RUN cd /src && \
  go build -ldflags="-s -w" -trimpath -o /majmun ./cmd/majmun

FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /majmun /majmun

USER 1337

ENTRYPOINT ["/majmun"]
