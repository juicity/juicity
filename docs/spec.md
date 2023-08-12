# Juicity 协议规范

Juicity 的设计理念为简约且能将 quic 发挥得当。

## 约定

协议规范中出现的关键词“必须”、“应当”、“不应”应按照 [RFC2119](https://datatracker.ietf.org/doc/html/rfc2119) 中的描述进行解释。

1. “必须”：对应 RFC2119 中的 “MUST”，意味着该定义是规范的绝对要求。
1. “应当”、“不应”：对应 RFC2119 中的 “SHOULD”、“SHOULD NOT”，意味着在特定情况下忽略该定义可能存在正当理由，但在选择不同的方式之前必须理解并仔细权衡其全部含义。

## 传输层

Juicity 使用的传输层是 quic，quic 保证了信息安全、多路复用、可靠性以及高带宽。

Juicity 要求 quic 至少支持 bbr 拥塞控制算法；要求 tls 的版本必须 1.3 以上， alpn 必须使用 h3。

需要注意的是，一般情况下，quic 对 maxOpenIncomingStreams 存在限制，客户端必须维护对端 quic connection 可用 stream 的动态数量，在可用数量不足时建立新的 connection 处理该 stream 的打开请求。当客户端不具有这样的能力时，例如 quic 底层库未暴露该接口时，客户端必须在同一个 connection 中打开累计 30 个 streams 后新建一个 connection 处理后续 stream 的打开。一般地，quic 底层库暴露的数量是准确的，从上层通过 close 和 open 来维护该数量是不准确的，实现不应通过这种方式维护该数量。

服务端的 maxOpenIncomingStreams 参数必须大于等于 30，客户端无此要求。

## 协议设计

### Authenticate

Juicity 通过 UUID 和 password 对用户进行鉴权。一个 quic connection 可以承载多个 streams，对于每个 quic connection，认证只需要一次即可。

与 Tuic 一样，Juicity 的认证发生在建立 connection 时，客户端打开一个 unistream 发送认证请求到服务端。

```p4
enum bit<8> CmdType {
    Authentication = 0;
};

header auth_h {
    bit<128> uuid;
    bit<256> token;
}

header_union command_body {
    auth_h auth;
};

header command_t {
    bit<8> version;
    CmdType cmd_type;
};
```

其中 version 为 0，token 使用如下方式产生：

> ExportKeyingMaterial returns length bytes of exported key material in a new slice as defined in RFC 5705.

```go
token = quicConnState.TLS.ExportKeyingMaterial(uuid, password, length=32)
```

认证请求没有应答。在客户端，代理请求无须也无法等待认证流程的结束，代理请求和认证请求可同时发送。在服务端，可同时接收认证请求和代理请求，但只有在认证成功后才开始处理代理请求，认证失败时则关闭整个 connection。

### Proxy

根据不同类型的四层协议，代理请求的行为略有不同，但代理头的格式是共享的。所有控制字段均为大端。

#### Header

代理头的具体格式如下：

```p4
enum bit<8> Network {
    TCP        = 1,
    UDP        = 3
};

enum bit<8> AddrType {
    IPV4       = 0,
    IPV6       = 1,
    DOMAIN     = 2
};

header domain_address_t {
    bit<2>      len;
    varbit<256> domain;
};

// address_t can be one of ipv4, ipv6 and domain.
header_union address_t {
    bit<32>          ipv4;
    bit<128>         ipv6;
    domain_address_t domain;
};

header proxy_t {
    Network   network;
    AddrType  addr_type;
    address_t address;
    bit<16>   port;
};
```

#### TCP

对于每个 TCP 连接（<sip, sport, tcp, dest, dport>）的代理请求，客户端打开一个 stream，发送代理头和荷载，其中 Network 为 TCP。一个承载 TCP 的 stream 只需要发送一次 proxy header。

Juicity 不解决长度混淆问题，因此代理头可单独发送，也可与数据字段一起发送。当具体实现选择与数据字段一起发送时，为了防止一些游戏场景的服务端推送出现问题，即服务端发送首包的场景，在客户端没有发送数据超过一定时间后，必须单独发送代理头，推荐值为 100ms 到 300ms。

#### UDP

Juicity 的 UDP 数据报基于 quic stream 传输，类似于 UDP over TCP。为了实现更好的 full-cone NAT，每一个源地址三元组（<sip, sport, UDP>）的数据报应当在同一个 stream 中传输，源地址三元组没有对应的 stream 时打开一个 stream。

由于每个 UDP 数据报均可指定不同的目的地址，因此对于每个 UDP 数据报的代理请求均要发送代理头和荷载，其中 Network 为 UDP。也即是说，在一个承载 UDP 的  stream 中可能会发送多次代理头。与 TCP 不同的是，荷载前需要给出 2 字节的荷载长度，如下：

```
[proxy header][len][payload]
```

在客户端维护源地址三元组与 stream 的映射，并通过 NAT timeout 控制映射和 stream 的生命周期。具体地，在没有收到和发送任何数据包超过 timeout 后，stream 可被关闭，并删除映射。

在服务端，一个 stream 对应一个 outbound UDP 端点，在 stream 关闭后删除 UDP 端点映射，服务端也可（MAY）建立 nat timeout 机制，在 timeout 后关闭 stream。在这种情况下，服务端的 nat timeout 应当大于建议值 3 分钟。

Juicity 的 UDP 支持 dial domain，服务端实现需要为每个承载 UDP 的 stream 建立一个域名到 ip 的映射，以便于在读出代理头的时候并其转为 IP 时保持映射稳定。

## 协议特点

Juicity 是基于 Tuic 的改进，主要改进 Tuic 的 UDP 所存在的一些问题。

1. 当 Tuic 的 UDP_relay_mode 使用 raw 时，在丢包线路中的应用层重试将变得严重，例如 DNS 的重试通常会发生在几秒后，较为影响体验。
1. 当 Tuic 的 UDP_relay_mode 使用 quic 时，每一个 UDP 数据报均使用单独的 unistream 传输，消耗不必要的资源。

Juicity 使用 UDP over Stream 解决这上述问题，并在规范中给出更多实现建议和约束，以避免其他可能出现的问题。
