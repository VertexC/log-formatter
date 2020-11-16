SHELL := /bin/bash

.PHONY: clean
clean:
	echo "a test"

## build: build main
.PHONY: build 
build:
	go build main.go

## go-test: go unit test
.PHONY: go-test
go-test: 
	go test -v ./...

.PHONY: services-start
services-start:
	docker-compose -f test/docker-compose.yml up -d

.PHONY: services-down
services-down:
	docker-compose -f test/docker-compose.yml down

.PHONY: file-file-test
file-file-test: build
	timeout --preserve-status 20s ./main -c test/file-file-test.yml
	@[ $(shell wc -l < output-test.txt) -eq $(shell wc -l < test/input-test.txt) ]
	rm output-test.txt

.PHONY: kafka-test
kafka-test: build
	$(MAKE) services-start
	sleep 10s
	timeout --preserve-status 20s ./main -c test/file-kafka-test.yml
	timeout --preserve-status 20s ./main -c test/kafka-file-test.yml
	@[ $(shell wc -l < output-test.txt) -eq $(shell wc -l < test/input-test.txt) ]
	$(MAKE) services-down
	rm output-test.txt

.PHONY: docker-push
docker-push-linux:
	GOOS=linux go build main.go
	docker build --tag log-formatter .
	docker tag log-formatter vertexc/log-formatter
	docker push vertexc/log-formatter:latest
