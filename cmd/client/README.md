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

- `listen` is the address that the socks5 and http server listen at. If you want authentication, write it like `user:pass@:1080`.
- Optional values of `congestion_control`: cubic, bbr, new_reno.
- `sni` can be omitted if domain is given in `server`.
- Set environment variable `QUIC_GO_ENABLE_GSO=true` to enable GSO, which can greatly improve the performance of sending and receiving packets. Notice that this option needs the support of NIC features. See more: <https://github.com/juicity/juicity/discussions/42>

## Run Options

| Name                  | Description                             | Default Value                | Required |
| --------------------- | --------------------------------------- | ---------------------------- | -------- |
| `--config, -c`        | specify config file path                | NA                           | yes      |
| `--disable-timestamp` | disable timestamp                       | false                        | no       |
| `--log-file`          | write logs to file                      | /var/log/juicity/juicity.log | no       |
| `--log-disable-color` | disable colorful log output             | false                        | no       |
| `--log-format`        | specify log format; options: [raw,json] | raw                          | no       |
