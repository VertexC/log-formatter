SHELL := /bin/bash

.PHONY: clean
clean:
	@-rm agents
	@-rm monitor-app

## build-sever: 
.PHONY: build-fe
build-fe:
	@rm -r build
	$(MAKE) -C ./agents-monitor all
	mkdir build && cp -r ./agents-monitor/dist build

.PHONY: build-agent
build-agent:
	GOOS=linux go build -o agent-app agent.go

.PHONY: build-monitor
build-monitor:
	GOOS=linux go build -o monitor-app monitor.go

## go-test: go unit test
.PHONY: go-test
go-test: 
	go test -v ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: services-start
services-start:
	docker-compose -f test/docker-compose.yml up -d

.PHONY: services-down
services-down:
	docker-compose -f test/docker-compose.yml down

.PHONY: file-file-test
file-file-test: build
	timeout --preserve-status 20s ./main -c test/file-file-test.yml
	@sh test/check-same-line.sh test/input-test.txt output-test.txt
	@rm output-test.txt

.PHONY: kafka-test
kafka-test: build
	$(MAKE) services-start
	sleep 10s
	timeout --preserve-status 20s ./main -c test/file-kafka-test.yml
	timeout --preserve-status 20s ./main -c test/kafka-file-test.yml
	@sh test/check-same-line.sh test/input-test.txt output-test.txt
	$(MAKE) services-down
	rm output-test.txt

.PHONY: docker-push-agent
docker-push-agent: | build-agent
	docker build --tag agent -f Dockerfile.agent .
	docker tag agent formatter/agent:$(docker_version)
	docker push formatter/agent

.PHONY: docker-push-agent
docker-push-agent: | build-monitor build-fe
	docker build --tag monitor -f Dockerfile.monitor .
	docker tag monitor formatter/monitor:$(docker_version)
	docker push formatter/monitor