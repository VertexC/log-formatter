package kafka

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"log"
	"sync"

	"github.com/VertexC/log-formatter/util"
)

type Consumer struct {
	ready   chan bool
	inputCh chan map[string]interface{}
	schema  string
}

type Config struct {
	Broker    string `yaml:"broker"`
	BatchSize int    `default:"1000" yaml:"batch_size"`
	GroupName string `default:"log-formatter" yaml:"group_name"`
	Topic     string `yaml:"topic"`
	Version   string `default:"2.4.0" yaml:"version"`
	Schema    string `yaml:"schema"`
}

var logger *util.Logger

func ExecuteGroup(config Config, inputCh chan map[string]interface{}) {

	logger = util.NewLogger("Kafka-Consumer-Group")

	sarama.Logger = logger.Trace

	version, err := sarama.ParseKafkaVersion(config.Version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}
	saramaCfg := sarama.NewConfig()
	// Adapt sarama version to Kafka version
	saramaCfg.Version = version

	// TODO: what this oldest parameter do?
	oldest := true
	if oldest {
		saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumer := Consumer{
		ready:   make(chan bool),
		inputCh: inputCh,
		schema:  config.Schema,
	}

	ctx, cancel := context.WithCancel(context.Background())
	brokers := []string{config.Broker}
	client, err := sarama.NewConsumerGroup(brokers, config.GroupName, saramaCfg)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			// TODO: add multi topic in config
			topics := []string{config.Topic}
			if err := client.Consume(ctx, topics, &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
	log.Println("Sarama consumer end!")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) decode(val []byte) map[string]interface{} {
	result := map[string]interface{}{}
	switch consumer.schema {
	case "json":
		err := json.Unmarshal(val, &result)
		if err != nil {
			logger.Error.Fatalf("Failed to parse json: %s\n", err)
		}
	case "":
		result["message"] = string(val)
	default:
		logger.Error.Fatalf("Invalid decode method: %s\n", consumer.schema)
	}
	return result
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		logger.Trace.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		consumer.inputCh <- consumer.decode(message.Value)
		session.MarkMessage(message, "")
	}

	return nil
}
