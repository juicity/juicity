# Get version from .git.
date=$(shell git log -1 --format="%cd" --date=short | sed s/-//g)
count=$(shell git rev-list --count HEAD)
commit=$(shell git rev-parse --short HEAD)
ifeq ($(wildcard .git/.),)
	VERSION ?= unstable-0.nogit
else
	VERSION ?= unstable-$(date).r$(count).$(commit)
endif
RUNTIME=$(shell go version)
ifndef CGO_ENALBED
    CGO_ENALBED := 0
endif

all: juicity-server juicity-client

juicity-server:
	CGO_ENALBED=$CGO_ENALBED go build -o $@ -trimpath \
		 -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" \
		 ./cmd/server

juicity-client:
	CGO_ENALBED=$CGO_ENALBED go build -o $@ -trimpath \
		 -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" \
		 ./cmd/client

.PHONY: juicity-server juicity-client all
