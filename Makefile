# Get version from .git.
date=$(shell git log -1 --format="%cd" --date=short | sed s/-//g)
count=$(shell git rev-list --count HEAD)
commit=$(shell git rev-parse --short HEAD)
ifeq ($(wildcard .git/.),)
	VERSION ?= unstable-0.nogit
else
	VERSION ?= unstable-$(date).r$(count).$(commit)
endif

juicity-server:
	go build -o $@ -trimpath -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" ./cmd/server

juicity-client:
	go build -o $@ -trimpath -ldflags "-s -w -X github.com/juicity/juicity/config.Version=$(VERSION)" ./cmd/client

.PHONY: juicity-server juicity-client
