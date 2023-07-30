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

- Optional values of `congestion_control`: cubic, bbr, new_reno.
- `sni` can be omitted if domain is in `server`.
