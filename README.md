# juicity

juicity is a quic-based proxy protocol, inspired by tuic.

In most cases, compared to tuic v5, juicity has following advantages:

1. More stable.
1. More actively maintained.
1. Better UDP implementation.
1. Better compatibility and consistency with clients in golang.

## [juicity-server](cmd/server)

## Link Format

Full parameters:
```
juicity://uuid:password@example.com:port?congestion_control=cubic&sni=example.com&allow_insecure=0
```

Mini parameters:
```
tuic://uuid:password@example.com:port?udp_relay_mode=native&congestion_control=cubic
```

## Clients

### Linux

- [daeuniverse/dae](https://github.com/daeuniverse/dae/pull/248).
