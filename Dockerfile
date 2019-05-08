FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /go/src/github.com/halkeye/discord-twitch-streamers
COPY . .
RUN set -ex && \
  go get github.com/ahmetb/govvv && \
  go get ./... && \
  GOOS=linux CGO_ENABLED=0 go build \
        -v -a -ldflags="-extldflags -static $(govvv -flags)" && \
  mv ./discord-twitch-streamers /usr/bin/

#FROM scratch
#FROM busybox:1.30
FROM alpine:3.9
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
COPY --from=builder /usr/bin/discord-twitch-streamers /app/discord-twitch-streamers
COPY ./static /app/static

# Set the binary as the entrypoint of the container
ENTRYPOINT [ "/app/discord-twitch-streamers" ]
