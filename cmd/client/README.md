# juicity-client

## Install

**Download from releases**

Multiple platforms and architectures are provited at <https://github.com/juicity/juicity/releases>.

**Build from sratch**

If you want to build from scratch:

```shell
git clone https://github.com/juicity/juicity
cd juicity
make juicity-client
```

## Run

```shell
./juicity-client run -c config.json
```

## Configuration

```json
{
    "listen": ":1080",
    "server": "<ip or domain>:<port>",
    "uuid": "00000000-0000-0000-0000-000000000000",
    "password": "my_password",
    "sni": "www.example.com",
    "allow_insecure": false,
    "congestion_control": "bbr",
    "log_level": "info"
}
```

- `listen` is the address that the socks5 and http server listen at.
- Optional values of `congestion_control`: cubic, bbr, new_reno.
- `sni` can be omitted if domain is given in `server`.
- Set environment variable `QUIC_GO_ENABLE_GSO=true` to enable GSO, which can greatly improve the performance of sending and receiving packets. Notice that this option needs the support of NIC features. See more: <https://github.com/orgs/juicity/discussions/42>
