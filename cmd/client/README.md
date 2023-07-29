# juicity-client

## Build

```shell
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
    "server": "112.32.62.11:23182",
    "uuid": "00000000-0000-0000-0000-000000000000",
    "password": "my_password",
    "sni": "www.example.com",
    "allow_insecure": false,
    "congestion_control": "bbr",
    "log_level": "info"
}
```

- Optional values of `congestion_control`: cubic, bbr, new_reno.
