FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /go/src/github.com/halkeye/discord-twitch-streamers
COPY . .
RUN set -ex && \
  go get ./... && \
  GOOS=linux CGO_ENABLED=0 go build \
        -v -a \
        -ldflags '-extldflags "-static"' && \
  mv ./discord-twitch-streamers /usr/bin/

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Retrieve the binary from the previous stage
COPY --from=builder /usr/bin/discord-twitch-streamers /usr/local/bin/discord-twitch-streamers
COPY ./static /usr/local/bin/static

# Set the binary as the entrypoint of the container
ENTRYPOINT [ "discord-twitch-streamers" ]
