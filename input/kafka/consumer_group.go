package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"sync"
)

type Consumer struct {
	ready   chan bool
	inputCh chan interface{}
}

func ExecuteGroup(config Config, inputCh chan interface{}, logFile string, verbose bool) {

	logger.Init(logFile, "Kafka-Consumer-Group", verbose)

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
	}

	// TODO: what does this do? golang context
	ctx, cancel := context.WithCancel(context.Background())
	// TODO: add multi brokers in config
	brokers := []string{config.Broker}
	client, err := sarama.NewConsumerGroup(brokers, config.GroupName, saramaCfg)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// defer wg.Done()
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

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		logger.Trace.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		consumer.inputCh <- map[string]interface{}{"message": string(message.Value)}
		session.MarkMessage(message, "")
	}

	return nil
}
