[![Build Status](https://travis-ci.org/VertexC/log-formatter.svg?branch=master)](https://travis-ci.org/VertexC/log-formatter)
[![Go Report Card](https://goreportcard.com/badge/github.com/VertexC/log-formatter)](https://goreportcard.com/report/github.com/VertexC/log-formatter)
# Log Formatter
Log Formatter provides configurable pipeline to process log data.

## Usage
```bash
-c string
    config file path (default "config.yml")
-cpuprof
    enable cpu profile
-memprof
    enable mem profile
-v    add TRACE/WARNING logging if enabled
```

## Documentation
https://godoc.org/github.com/VertexC/log-formatter

## Configuration
**A Example Config**: Simply forward lines from one file to another.
```yaml
log: "logs"
output:
  target: "file"
  file: "input.txt"
input:
  target: "file"
  file: "output.txt"
formatter: 
  type: ""
```
The config can be modulized with `!include`
```yaml
log: "logs"
output: !include modules/output.yml
input: !include modules/input.yml
formatter: !include modules/formatter.yml
```
More templates are available under `config.modules`, please checkout.

## Docker
The lastest built docker will be pushed and available on [docker hub](https://hub.docker.com/repository/docker/vertexc/log-formatter)