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
juicity://uuid:password@122.12.31.66:port?congestion_control=cubic&sni=www.example.com&allow_insecure=0
```

Mini parameters:
```
juicity://uuid:password@example.com:port?congestion_control=cubic
```

## Clients

- [juicity/juicity-client](cmd/client).
- [daeuniverse/dae](https://github.com/daeuniverse/dae/pull/248) (Linux Only).
