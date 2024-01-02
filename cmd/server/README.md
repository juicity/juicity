# juicity-server

## Install

**Download from releases**

Multiple platforms and architectures are provited at <https://github.com/juicity/juicity/releases>.

**Build from sratch**

If you want to build from scratch:

```shell
git clone https://github.com/juicity/juicity
cd juicity
make CGO_ENABLED=0 juicity-server
```

## Run

```shell
./juicity-server run -c config.json
```

Or with Docker

```
docker run --name juicity \
  --restart always \
  --network host \
  -v /path/to/config.json:/etc/juicity/server.json \
  -v /path/to/fullchain.cer:/path/to/fullchain.cer \
  -v /path/to/private.key:/path/to/private.key \
  -dit ghcr.io/juicity/juicity:main
```

## Configuration

Mini configuration:

```json
{
  "listen": ":23182",
  "users": {
    "00000000-0000-0000-0000-000000000000": "my_password"
  },
  "certificate": "/path/to/fullchain.cer",
  "private_key": "/path/to/private.key",
  "congestion_control": "bbr",
  "disable_outbound_udp443": true,
  "log_level": "info"
}
```

Full configuration:

```json
{
  "listen": ":23182",
  "users": {
    "00000000-0000-0000-0000-000000000000": "my_password"
  },
  "certificate": "/path/to/fullchain.cer",
  "private_key": "/path/to/private.key",
  "congestion_control": "bbr",
  "log_level": "info",
  "fwmark": "0x1000",
  "send_through": "113.25.132.3",
  "dialer_link": "socks5://127.0.0.1:1080",
  "disable_outbound_udp443": true
}
```

- `congestion_control`: one of cubic, bbr, new_reno.
- `fwmark` is useful for iptables/nft.
- `send_through` is the interface IP to specify to use.
- `dialer_link` can be extreme flexible. Juicity support many protocols, even proxy chains. See [proxy-protocols](https://github.com/daeuniverse/dae/blob/main/docs/en/proxy-protocols.md) [中文](https://github.com/daeuniverse/dae/blob/main/docs/zh/proxy-protocols.md).
- `disable_outbound_udp443`: usually quic traffic. Suggest to disable it because quic usually consumes too much cpu/mem resources.

## Arguments

Run `juicity-server run -h` to get the full arguments.

## UUID Generator

You may make use of an [online uuid-generator](https://www.v2fly.org/en_US/awesome/tools.html) from [@v2fly](https://github.com/v2fly) to generate a legitimate uuid.

Alternatively, for system that ships with Python (e.g Debian or Ubuntu), you may use the following commands to generate a UUID

```bash
python3 -c "from uuid import uuid4;print(uuid4())"
```

Or install a cross-platform binary `uuidgen`:

```bash
# e.g debian
sudo apt install uuid-runtime

# usage
uuidgen
```

Also see [#63](https://github.com/juicity/juicity/issues/63)

## Generate ShareLink

```bash
juicity-server generate-sharelink -c /etc/juicity/server.json
# output
juicity://00000000-0000-0000-0000-000000000000:mypassword@1.2.3.4:15333?congestion_control=bbr&pinned_certchain_sha256=5ykL73pOK7NAu92A48dCrFjDqDowdChUSmlpQzudmvc%3D&sni=example.com
```
