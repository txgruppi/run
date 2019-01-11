![Codeship](https://img.shields.io/codeship/cb3a7670-f7ee-0136-66a3-16fab027ee75.svg?style=flat-square)
[![Codecov](https://img.shields.io/codecov/c/github/txgruppi/run.svg?style=flat-square)](https://codecov.io/github/txgruppi/run)
[![Go Report Card](https://img.shields.io/badge/go_report-A+-brightgreen.svg?style=flat-square)](https://goreportcard.com/report/github.com/txgruppi/run)

# run

`run` replaces _tokens_ in a config file template by the values of _environment variables_ with the same as as the tokens, saves everything in a new config file and it executes a command.

It was designed to be used in _docker containers_ where a config file should receive values from the _environment variables_ before running the container's command.

## Usage

The example below is of a container with a _webserver_ but before starting the server it will compile the config file template using the `run` command.

### Environment variables (.env)

```
MONGO_URL="mongodb://user:password@my.server.com/mydb"
JWT_SECRET="my$uper$ecret"
SERVER_BIND="0.0.0.0"
SERVER_PORT="8000"
```

### Config template (config.toml.dist)

```
[database]
url = "{{MONGO_URL}}"

[jwt]
secret = "{{JWT_SECRET}}"

[server]
bind = "{{SERVER_BIND}}"
port = "{{SERVER_PORT}}"
```

### Dockerfile (txgruppi/run-sample)

```
FROM busybox:1.25.1

MAINTAINER Tarcisio Gruppi <txgruppi@gmail.com>

ADD https://github.com/txgruppi/run/releases/download/0.0.1/run_linux_amd64 /app/run
ADD ./config.toml.dist /app/config.toml.dist
ADD ./server /app/server

RUN run -i /app/config.toml.dist -o /app/config.toml /app/server -c /app/config.toml
```

### Running the container

```
docker run -d --restart=always --env-file .env -p 8000 txgruppi/run-sample
```

### Compiled config file (config.toml)

```
[database]
url = "mongodb://user:password@my.server.com/mydb"

[jwt]
secret = "my$uper$ecret"

[server]
bind = "0.0.0.0"
port = "8000"
```
