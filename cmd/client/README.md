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

Mini configuration:

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

Full configuration:

```json
{
  "listen": ":1080",
  "server": "<ip or domain>:<port>",
  "uuid": "00000000-0000-0000-0000-000000000000",
  "password": "my_password",
  "sni": "www.example.com",
  "allow_insecure": false,
  "congestion_control": "bbr",
  "pinned_certchain_sha256": "aQc4fdF4Nh1PD6MsCB3eofRyfRz5R8jJ1afgr37ABZs=",
  "log_level": "info"
}
```

- `listen` is the address that the socks5 and http server listen at. If you want authentication, write it like `user:pass@:1080`.
- Optional values of `congestion_control`: cubic, bbr, new_reno.
- `sni` can be omitted if domain is given in `server`.
- `pinned_certchain_sha256` is the pinned hash of remote TLS certificate chain. You can generate it by `juicity-server generate-certchain-hash [fullchain_cert_file]`. See <https://github.com/juicity/juicity/issues/34>.
- Set environment variable `QUIC_GO_ENABLE_GSO=true` to enable GSO, which can greatly improve the performance of sending and receiving packets. Notice that this option needs the support of NIC features. See more: <https://github.com/juicity/juicity/discussions/42>

## Arguments

| Name                  | Description                                                                                                 | Default Value          | Required |
| --------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------- | -------- |
| `--config, -c`        | specify config file path                                                                                    | NA                     | yes      |
| `--disable-timestamp` | disable timestamp                                                                                           | false                  | no       |
| `--log-file`          | write logs to file                                                                                          | "" (not write to file) | no       |
| `--log-format`        | specify log format; options: [raw,json]                                                                     | raw                    | no       |
| `--log-disable-color` | disable colorful log output                                                                                 | false                  | no       |
| `--log-max-size`      | specify maximum size in megabytes of the log file before it gets rotated                                    | 10 megabytes           | no       |
| `--log-max-backups`   | specify the maximum number of old log files to retain                                                       | 1 copy                 | no       |
| `--log-max-age`       | specify the maximum number of days to retain old log files based on the timestamp encoded in their filename | 1 day                  | no       |
| `--log-compress`      | enable log compression                                                                                      | true                   | no       |
