# Juicity Protocol Specification

[**简体中文**](./spec.md) | [**English**](./spec_en.md)

The design philosophy of Juicity is to be simple yet effective in utilizing QUIC.

## Conventions

The key terms "MUST", "SHOULD" and "SHOULD NOT" in this protocol specification are to be interpreted as described in [RFC2119](https://datatracker.ietf.org/doc/html/rfc2119).

The term "connection" in the protocol specification generally refers to QUIC connection. In Juicity, TCP connections are carried over streams of QUIC connections, thus:
```
  QUIC connection : QUIC stream : TCP connection
= 1               : N           : N
```

## Transport Layer

Juicity uses QUIC as its transport layer, which ensures information security, multiplexing, reliability, and high bandwidth.

Juicity requires that QUIC MUST support the BBR congestion control algorithm at a minimum. It also requires that the version of TLS MUST be 1.3 or above, and ALPN MUST be h3.

It's important to note that under normal circumstances, QUIC may impose limits on `maxOpenIncomingStreams`. The client MUST maintain a dynamic count of available streams on the remote QUIC connection. When the available count is insufficient, new connections MUST be established to handle incoming stream open requests. If the client lacks the capability, for instance when the underlying QUIC library doesn't expose such an interface, the client MUST create a new QUIC connection after opening a cumulative total of 30 streams on the same connection to handle subsequent stream opens. Generally, the quantity exposed by the QUIC library is accurate, and maintaining this count inaccurately through close and open counts from the upper layer SHOULD NOT be implemented.

The `maxOpenIncomingStreams` parameter on the server side MUST be greater than or equal to 30, while the client has no such requirement.

## Protocol Design

### Authenticate

Juicity performs user authentication using UUID and password. A single QUIC connection can carry multiple QUIC streams, and for each QUIC connection, authentication is required only once.

Similar to Tuic, Juicity's authentication occurs during the establishment of the QUIC connection. The client opens a unidirectional stream to send an authentication request to the server.

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

Where the version is 0, and the token is generated as follows:

> ExportKeyingMaterial returns length bytes of exported key material in a new slice as defined in RFC 5705.

```go
token = quicConnState.TLS.ExportKeyingMaterial(uuid, password, length=32)
```

The authentication request does not receive a response. On the client side, proxy requests need not and cannot wait for the authentication process to finish. Proxy requests and authentication requests can be sent simultaneously. On the server side, however, only after successful authentication will the server start processing proxy requests; if authentication fails, the entire QUIC connection should be closed.

### Proxy

Proxy requests have slightly different behaviors based on different types of Layer 4 protocols, but they share the same proxy header format. All control fields are in big-endian format.

#### Header

The specific format of the proxy header is as follows:

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
    bit<8>       len;
    varbit<2048> domain;
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

For each proxy request of a TCP connection (<source IP, source port, TCP, destination, destination port>), the client opens a stream and sends the proxy header and payload, where Network is TCP. A stream carrying TCP only needs to send the proxy header once.

Juicity does not address the problem of length confusion. Therefore, the proxy header can be sent separately or together with the data field. When the specific implementation chooses to send it with the data field, to prevent issues with server-side pushes in certain gaming scenarios, such as when the server sends the first packet, the client MUST send the proxy header separately if no data has been sent for a certain period, preferably between 100ms and 300ms.

#### UDP

Juicity's UDP datagrams are transmitted over QUIC streams, similar to UDP over TCP. To achieve better full-cone NAT support, the datagrams of each source address triplet (<source IP, source port, UDP>) SHOULD be transmitted over the same stream. If there's no corresponding stream for a source address triplet, a new stream should be opened.

As each UDP datagram can have a different destination address, proxy headers and payloads MUST be sent for each UDP datagram, where Network is UDP. Unlike TCP, a 2-byte payload length needs to be provided before the payload, as shown below:

```
[proxy header][length][payload]
```

On the client side, the mapping between source address triplets and streams is maintained, and the mapping and stream lifecycle are controlled by NAT timeout. Specifically, if no data packets are sent or received for a period beyond the timeout, the stream can be closed, and the mapping is deleted.

On the server side, a stream corresponds to an outbound UDP endpoint. After the stream is closed, the mapping of the UDP endpoint is deleted. The server MAY also establish a NAT timeout mechanism and close the stream after the timeout. In this case, the server's NAT timeout SHOULD be greater than the recommended value of 3 minutes.

Juicity's UDP also supports dialing domains. The server implementation needs to establish a mapping from domain to IP for each stream carrying UDP to convert the domain to an IP when reading the proxy header, thus maintaining a stable mapping.

## Protocol Features

Juicity is an improvement over Tuic and addresses certain issues in Tuic's UDP handling.

1. When Tuic's udp_relay_mode is set to native, application-level retries in case of packet loss become severe. For example, DNS retries often occur after a few seconds, affecting user experience.
2. When Tuic's udp_relay_mode is set to quic, each UDP datagram is transmitted over a separate unidirectional stream, resulting in unnecessary resource consumption.

Juicity uses UDP over Stream to address these issues and provides more implementation suggestions and constraints in the specification to avoid potential problems.
