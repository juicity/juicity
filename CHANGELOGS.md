# Changelogs

Also seen in [GitHub Releases](https://github.com/juicity/juicity/releases)

## Query history releases

```bash
curl --silent "https://api.github.com/repos/juicity/juicity/releases" | jq -r '.[] | {tag_name,created_at,release}'
```

## Releases

<!-- BEGIN NEW TOC ENTRY -->

- [v0.1.0 (latest)](#v010-latest)
<!-- BEGIN NEW CHANGELOGS -->

### v0.1.0 (latest)

> Release date: 2023/07/30

### Notes

> **Note**: initial release

Juicity is a quic-based proxy protocol. It has strong performance and is of great help to improve the network quality of the proxy. We have mature experience in proxy protocol, which can ensure that you avoid procedural and design problems as much as possible when using them. Have fun!

Juicity 是一个基于 quic 的代理协议，它有着强劲的性能，对网络质量较差的代理环境有较大的改善。我们在代理协议上具有成熟的经验，能保证您在使用时尽可能避免程序性和设计性的问题。最后，玩得开心！
