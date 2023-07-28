juice-server:
	go build -o $@ -trimpath -ldflags "-s -w" ./cmd/server

.PHONY: juice-server
