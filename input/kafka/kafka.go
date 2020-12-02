package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"sync"

	"github.com/VertexC/log-formatter/util"
)

type Consumer struct {
	ready  chan bool
	docCh  chan util.Doc
	schema string
	logger *util.Logger
}

type KafkaConfig struct {
	Brokers   []string `yaml:"brokers"`
	BatchSize int      `default:"1000" yaml:"batch_size"`
	GroupName string   `default:"log-formatter" yaml:"group_name"`
	Topic     string   `yaml:"topic"`
	Version   string   `default:"2.4.0" yaml:"version"`
	Schema    string   `yaml:"schema"`
	// Worker is the number of workers in sarama
	Worker int `yaml:"worker"`
}

type worker struct {
	consumer *Consumer
	client   sarama.ConsumerGroup
	topic    string
	logger   *util.Logger
}

type KafkaInput struct {
	logger  *util.Logger
	workers []*worker
	config  KafkaConfig
}

func NewKafkaInput(config KafkaConfig, docCh chan util.Doc) *KafkaInput {
	logger := util.NewLogger("kafka-consumer")

	sarama.Logger = logger.Trace
	version, err := sarama.ParseKafkaVersion(config.Version)
	if err != nil {
		log.Fatalf("Error parsing Kafka version: %v", err)
	}

	input := &KafkaInput{
		logger:  logger,
		config:  config,
		workers: []*worker{},
	}
	if config.Worker == 0 {
		config.Worker = 1
	}
	for i := 0; i < config.Worker; i++ {
		consumer := &Consumer{
			ready:  make(chan bool),
			docCh:  docCh,
			schema: config.Schema,
			logger: logger,
		}
		// create new saram config for different clientid
		saramaCfg := sarama.NewConfig()
		saramaCfg.ClientID = fmt.Sprintf("saram%d", i)
		// Adapt sarama version to Kafka version
		saramaCfg.Version = version

		// TODO: what this oldest parameter do?
		oldest := true
		if oldest {
			saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
		}

		client, err := sarama.NewConsumerGroup(config.Brokers, config.GroupName, saramaCfg)
		if err != nil {
			log.Panicf("Error creating consumer group client: %v", err)
		}
		input.workers = append(input.workers,
			&worker{
				consumer: consumer,
				client:   client,
				topic:    config.Topic,
				logger:   input.logger,
			})
	}
	return input
}

func (input *KafkaInput) Run() {
	for _, worker := range input.workers {
		go worker.Run()
	}
}

func (w *worker) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			topics := []string{w.topic}
			if err := w.client.Consume(ctx, topics, w.consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			w.consumer.ready = make(chan bool)
		}
	}()

	<-w.consumer.ready // Await till the consumer has been set up
	w.logger.Info.Println("Sarama consumer up and running!...")

	select {
	case <-ctx.Done():
		w.logger.Info.Println("terminating: context cancelled")
	}
	cancel()
	wg.Wait()
	if err := w.client.Close(); err != nil {
		w.logger.Warning.Printf("Error closing client: %v", err)
	}
	w.logger.Info.Println("Sarama consumer end!")
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

func (consumer *Consumer) decode(val []byte) util.Doc {
	result := util.Doc{}
	switch consumer.schema {
	case "json":
		err := json.Unmarshal(val, &result)
		if err != nil {
			consumer.logger.Error.Fatalf("Failed to parse json: %s\n", err)
		}
	case "":
		result["message"] = string(val)
	default:
		consumer.logger.Error.Fatalf("Invalid decode method: %s\n", consumer.schema)
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
		consumer.logger.Trace.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		consumer.docCh <- consumer.decode(message.Value)
		session.MarkMessage(message, "")
	}

	return nil
}
