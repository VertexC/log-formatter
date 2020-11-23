[![Build Status](https://travis-ci.org/VertexC/log-formatter.svg?branch=master)](https://travis-ci.org/VertexC/log-formatter)
[![Go Report Card](https://goreportcard.com/badge/github.com/VertexC/log-formatter)](https://goreportcard.com/report/github.com/VertexC/log-formatter)
[![codecov](https://codecov.io/gh/VertexC/log-formatter/branch/master/graph/badge.svg?token=ULNP7LB4AI)](https://codecov.io/gh/VertexC/log-formatter)
# Log Formatter
Log Formatter provides configurable pipeline to process log data.

## Usage
### build from source
```bash
bash$ go get github.com/VertexC/log-formatter
bash$ cd $GOPATH/github.com/VertexC/log-formatter
bash$ go build main.go
bash$ ./main -help
Usage of ./main:
  -c string
        config file path (default "config.yml")
  -cpuprof
        enable cpu profile
  -memprof
        enable mem profile
  -v    add TRACE/WARNING logging if enabled
```

### docker
Docker images are available on [docker hub](https://hub.docker.com/r/vertexc/log-formatter/tags), with branch name as tag (`master` is tagged as `latest`).


The docker image is built without any entries point. The executable binary is `/app/log-formatter`.
```bash
docker run -i -a stdin -a stdout -a stderr -v <local-config.yml>:/app/config.yml vertexc/log-formatter /app/log-formatter -h
```
## Documentation
https://godoc.org/github.com/VertexC/log-formatter

## Config File
The config can be modulized with `!include`
```yaml
log: "logs"
output: !include modules/output.yml
input: !include modules/input.yml
formatter: !include modules/formatter.yml
```
More templates are available under [modules](./config.modules/), please checkout.