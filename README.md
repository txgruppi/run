![Codeship](https://img.shields.io/codeship/cb3a7670-f7ee-0136-66a3-16fab027ee75.svg?style=flat-square)
[![Codecov](https://img.shields.io/codecov/c/github/txgruppi/run.svg?style=flat-square)](https://codecov.io/github/txgruppi/run)
[![Go Report Card](https://img.shields.io/badge/go_report-A+-brightgreen.svg?style=flat-square)](https://goreportcard.com/report/github.com/txgruppi/run)

# run

`run` replaces _tokens_ in a config file tempalte by values from the specific data sources, saves a new config file and executes a command.

It was designed to be used in _docker containers_ where a config file should receive values from the data sources before running the container's command.

## Data sources

- Environment variables
- Local JSON file
- Remote JSON file

## Options

```
--input value, -i value        The config template with the tokens to be replaced [$RUN_INPUT]
--output value, -o value       The output path for the compiled config file [$RUN_OUTPUT]
--delay value, -d value        Number of seconds to wait before running the command (default: 0) [$RUN_DELAY]
--json value, -j value         JSON data to be used by JSONLoader [$RUN_JSON]
--remote-json value, -r value  URL to a JSON file to be used by RemoteJSONLoader [$RUN_REMOTE_JSON]
--json-file value, -f value    Path to a JSON file to be used by JSONFileLoader [$RUN_JSON_FILE]
--aws-secret value             The ARN or name of a secret with a JSON encoded value [$RUN_AWS_SECRET_ARN]
--help, -h                     show help
--version, -v                  print the version
```

## Example

The example below is of a container with a _webserver_ but before starting the server it will compile the config file template using the `run` command.

### Environment variables (.env)

```shell
MONGO_URL="mongodb://user:password@my.server.com/mydb"
JWT_SECRET="my$uper$ecret"
SERVER_BIND="0.0.0.0"
SERVER_PORT="8000"
```

### Local JSON file (/mnt/shared/secrets/vars.json)

```json
{
  "server": {
    "bind": "0.0.0.0"
  }
}
```

### Remote JSON file (http://config-service/app/config.json)

```json
{
  "server": {
    "port": "1234"
  }
}
```

### AWS SecretManager

Secret name: `jwtconfig`

```
{
  "jwt": {
    "secret": "myjwtsecret"
  }
}
```

### Config template (config.toml.dist)

```toml
[database]
url = "{{MONGO_URL}}"

[jwt]
secret = "{{jwt.secret|JWT_SECRET}}"

[server]
bind = "{{server.bind|SERVER_BIND}}"
port = "{{server.port|SERVER_PORT}}"
```

### Dockerfile

```dockerfile
FROM busybox:1.25.1

MAINTAINER Tarcisio Gruppi <txgruppi@gmail.com>

ADD https://github.com/txgruppi/run/releases/download/0.0.1/run_linux_amd64 /app/run
ADD ./config.toml.dist /app/config.toml.dist
ADD ./server /app/server

RUN run \
  -d 2 \
  --json-file /mnt/shared/secrets/vars.json \
  --remote-json http://config-service/app/config.json \
  --aws-secret jwtconfig \
  -i /app/config.toml.dist \
  -o /app/config.toml \
  /app/server -c /app/config.toml
```

### Running the container

```shell
docker run -d --restart=always --env-file .env -p 1234 txgruppi/run-sample
```

### Compiled config file (config.toml)

```toml
[database]
url = "mongodb://user:password@my.server.com/mydb"

[jwt]
secret = "myjwtsecret"

[server]
bind = "0.0.0.0"
port = "1234"
```
