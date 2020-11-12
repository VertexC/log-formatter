clean:
	echo "a test"

build:
	go build main.go

## go-test: go unit test
.PHONY: go-test
go-test: 
	go test -v ./...

## kafka-consumer-test: consume from kafka and forward message to console
.PHONY: kafka-consumer-test
kafka-consumer-test: build
	 timeout --preserve-status 20s ./main -c test/kafka-console-test.yml
