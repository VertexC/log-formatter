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

services-start:
	docker-compose -f test/docker-compose.yml up -d

services-down:
	docker-compose -f test/docker-compose.yml down

## kafka-consumer-test: consume from kafka and forward message to console
.PHONY: kafka-consumer-test
kafka-consumer-test: build
	timeout --preserve-status 20s ./main -c test/kafka-console-test.yml

pipeline-test:
	$(MAKE) kafka-consumer-test

local-test:
	@echo "======= start go unit test ======"
	$(MAKE) go-test
	@echo "======= start pipeline test ======"
	$(MAKE) services-start
	$(MAKE) pipeline-test
	$(MAKE) services-down
