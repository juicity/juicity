# juicity

<p align="left">
    <img src="https://github.com/juicity/juicity/actions/workflows/build.yml/badge.svg" alt="Build"/>
    <img src="https://custom-icon-badges.herokuapp.com/github/license/juicity/juicity?logo=law&color=blue" alt="License"/>
    <img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fjuicity%2Fjuicity&count_bg=%23493DC8&title_bg=%23555555&icon=&icon_color=%23E7E7E7&title=hits&edge_flat=false"/>
    <img src="https://custom-icon-badges.herokuapp.com/github/v/release/juicity/juicity?logo=rocket" alt="version">
    <img src="https://custom-icon-badges.herokuapp.com/github/issues-pr-closed/juicity/juicity?color=purple&logo=git-pull-request&logoColor=white"/>
    <img src="https://custom-icon-badges.herokuapp.com/github/last-commit/juicity/juicity?logo=history&logoColor=white" alt="lastcommit"/>
</p>

juicity is a quic-based proxy protocol, inspired by tuic.

In most cases, compared to tuic v5, juicity has following advantages:

- [x] More stable.
- [x] More actively maintained.
- [x] Better UDP implementation.
- [x] Better compatibility and consistency with clients in golang.

## [juicity-server](cmd/server)

## Link Format

Full parameters:

```shell
juicity://uuid:password@122.12.31.66:port?congestion_control=cubic&sni=www.example.com&allow_insecure=0
```

Mini parameters:

```shell
juicity://uuid:password@example.com:port?congestion_control=cubic
```

## Clients

- [juicity/juicity-client](cmd/client) (Official)
- [daeuniverse/dae](https://github.com/daeuniverse/dae/pull/248) (Official, Linux Only)

## License

[AGPL-3.0 (C) juicity](https://github.com/juicity/juicity/blob/main/LICENSE)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/juicity/juicity.svg)](https://starchart.cc/juicity/juicity)
