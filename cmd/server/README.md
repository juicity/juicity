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
  "fwmark": "0x1000",
  "send_through": "113.25.132.3",
  "log_level": "info"
}
```

- Optional values of `congestion_control`: cubic, bbr, new_reno.
- `fwmark` is useful for iptables/nft.
- `send_through` is the interface IP to specify to use.
- Set environment variable `QUIC_GO_ENABLE_GSO=true` to enable GSO, which can greatly improve the performance of sending and receiving packets. Notice that this option needs the support of NIC features. See more: <https://github.com/juicity/juicity/discussions/42>

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
