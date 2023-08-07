FROM golang:1.20-alpine AS builder
COPY . /go/src/github.com/juicity/juicity
WORKDIR /go/src/github.com/juicity/juicity
ENV CGO_ENABLED=0
RUN apk add --no-cache make
RUN make juicity-server

FROM alpine AS dist
RUN set -ex \
    && apk upgrade \
    && apk add bash tzdata ca-certificates \
    && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/github.com/juicity/juicity/juicity-server /usr/local/bin/juicity-server
ENTRYPOINT ["juicity-server", "run", "-c", "/etc/juicity/config.json"]