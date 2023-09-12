# Get version from .git.
date=$(shell git log -1 --format="%cd" --date=short | sed s/-//g)
count=$(shell git rev-list --count HEAD)
commit=$(shell git rev-parse --short HEAD)

ifeq ($(wildcard .git/.),)
	VERSION ?= unstable-0.nogit
else
	VERSION ?= unstable-$(date).r$(count).$(commit)
endif

ifndef CGO_ENABLED
	CGO_ENABLED := 0
endif

all: juicity-server juicity-client

juicity-server:
	CGO_ENABLED=$(CGO_ENABLED) go build -o $@ -trimpath \
		 -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" \
		 ./cmd/server

juicity-client:
	CGO_ENABLED=$(CGO_ENABLED) go build -o $@ -trimpath \
		 -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" \
		 ./cmd/client

.PHONY: juicity-server juicity-client all
