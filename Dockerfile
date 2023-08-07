ARG APP_DIR=/app

### Builder ###
FROM golang:1.20-alpine AS builder

ARG APP_DIR

RUN apk add --no-cache make git

WORKDIR ${APP_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN make juicity-server

### Prod ###
FROM alpine:latest AS dist

ARG APP_DIR

RUN set -ex \
    && apk upgrade \
    && apk add tzdata ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder ${APP_DIR}/juicity-server /usr/bin/juicity-server
CMD ["juicity-server", "run", "-c", "/etc/juicity/server.json"]
