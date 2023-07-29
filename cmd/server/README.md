# juicity-server

## Build

```shell
make CGO_ENABLED=0 juicity-server
```

## Run

```shell
./juicity-server run -c config.json
```

## Configuration

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

- Optional values of `congestion_control`: cubic, bbr, new_reno.
