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

.PHONY: file-file-test
file-file-test: build
	timeout --preserve-status 20s ./main -c test/file-file-test.yml
	@[ $(shell wc -l < output-test.txt) -eq $(shell wc -l < test/input-test.txt) ]
	rm output-test.txt

.PHONY: pipeline-test
pipeline-test:
	$(MAKE) services-start
	sleep 10s
	$(MAKE) kafka-consumer-test
	$(MAKE) file-file-test
	$(MAKE) services-down

.PHONY: local-test
local-test:
	@echo "======= start go unit test ======"
	$(MAKE) go-test
	@echo "======= start pipeline test ======"
	$(MAKE) pipeline-test
