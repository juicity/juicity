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

Min configuration:

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

## Run Options

| Name                  | Description                             | Default Value                | Required |
| --------------------- | --------------------------------------- | ---------------------------- | -------- |
| `--config, -c`        | specify config file path                | NA                           | yes      |
| `--disable-timestamp` | disable timestamp                       | false                        | no       |
| `--log-file`          | write logs to file                      | /var/log/juicity/juicity.log | no       |
| `--log-disable-color` | disable colorful log output             | false                        | no       |
| `--log-format`        | specify log format; options: [raw,json] | raw                          | no       |
