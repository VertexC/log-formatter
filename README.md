[![Build Status](https://travis-ci.org/VertexC/log-formatter.svg?branch=master)](https://travis-ci.org/VertexC/log-formatter)
[![Go Report Card](https://goreportcard.com/badge/github.com/VertexC/log-formatter)](https://goreportcard.com/report/github.com/VertexC/log-formatter)
[![codecov](https://codecov.io/gh/VertexC/log-formatter/branch/master/graph/badge.svg?token=ULNP7LB4AI)](https://codecov.io/gh/VertexC/log-formatter)
# Log Formatter: Logstash in Golang
**Log Formatter** is a **light-weight**, **extensible** and **production-ready** framework in golang to process log data like [Logstash](https://github.com/elastic/logstash). It ingests data from `input` as documents, then each document is processed (filter/drop/enhance) by `pipeline`, and finally sent to `output`.

## Usage and Example
### Standalone Agent
### With Monitor

## Docker
Docker images of `agent` and `monitor` are available on [docker hub](https://hub.docker.com/r/formatter).

## K8S Deployment Example