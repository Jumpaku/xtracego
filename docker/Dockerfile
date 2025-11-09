# Dockerfile for building the image
FROM golang:1.25.0-alpine3.22 AS builder

ARG XTRACEGO_VERSION=latest
RUN go install "github.com/Jumpaku/xtracego/cmd/xtracego@$XTRACEGO_VERSION"


FROM golang:1.25.0-alpine3.22

COPY --from=builder /go/bin/xtracego /bin/xtracego
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
WORKDIR /workspace
ENTRYPOINT ["/docker-entrypoint.sh"]
