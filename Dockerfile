ARG APP_DIR=/app

### Builder ###
FROM golang:1.20-alpine AS builder

ARG APP_DIR

RUN apk add --no-cache make

WORKDIR ${APP_DIR}
COPY . .

ENV CGO_ENABLED=0
RUN make juicity-server

### Prod ###
FROM alpine:latest AS dist

ARG APP_DIR

RUN set -ex \
    && apk upgrade \
    && apk add bash tzdata ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=builder ${APP_DIR}/juicity-server /usr/bin/juicity-server
CMD ["juicity-server", "run", "-c", "/etc/juicity/config.json"]
