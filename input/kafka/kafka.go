package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/input"
	"github.com/VertexC/log-formatter/util"

	"github.com/Shopify/sarama"
	"gopkg.in/yaml.v3"
)

type Consumer struct {
	ready  chan bool
	docCh  chan map[string]interface{}
	schema string
	logger *util.Logger
}

type KafkaConfig struct {
	Base      config.ConfigBase
	Brokers   []string `yaml:"brokers"`
	GroupName string   `yaml:"group_name"`
	Topic     string   `yaml:"topic"`
	Version   string   `yaml:"version"`
	Schema    string   `yaml:"schema"`
	// Worker is the number of workers in sarama
	Worker int `yaml:"worker"`
}

type worker struct {
	consumer *Consumer
	client   sarama.ConsumerGroup
	topic    string
	logger   *util.Logger
	docCh    chan map[string]interface{}
}

type KafkaInput struct {
	logger  *util.Logger
	workers []*worker
	config  KafkaConfig
	docCh   chan map[string]interface{}
}

func init() {
	input.Register("kafka", NewKafkaInput)
}

func NewKafkaInput(content interface{}) (input.Input, error) {
	configMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get mapStr from config")
	}

	data, err := yaml.Marshal(&configMapStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to process given content as yaml: %s", err)
	}

	// TODO: can we avoid such manual mandantory fields settings by using struct's tags?
	config := KafkaConfig{
		Base: config.ConfigBase{
			Content:          configMapStr,
			MandantoryFields: []string{"brokers", "topic"},
		},
		GroupName: "log-formatter",
		Version:   "2.4.0",
		Schema:    "",
		Worker:    1,
	}

	yaml.Unmarshal(data, &config)

	logger := util.NewLogger("INPUT_KAFKA")

	sarama.Logger = logger.Trace
	version, err := sarama.ParseKafkaVersion(config.Version)
	if err != nil {
		log.Fatalf("Error parsing Kafka version: %v", err)
	}

	input := &KafkaInput{
		logger:  logger,
		config:  config,
		workers: []*worker{},
		docCh:   make(chan map[string]interface{}, 1000),
	}
	if config.Worker == 0 {
		config.Worker = 1
	}

	for i := 0; i < config.Worker; i++ {
		consumer := &Consumer{
			ready:  make(chan bool),
			docCh:  input.docCh,
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

		// FIXME:
		client, err := sarama.NewConsumerGroup(config.Brokers, config.GroupName, saramaCfg)
		if err != nil {
			return nil, fmt.Errorf("Error creating consumer group client: %v", err)
		}
		input.workers = append(input.workers,
			&worker{
				consumer: consumer,
				client:   client,
				topic:    config.Topic,
				logger:   input.logger,
			})
	}
	return input, nil
}

func (input *KafkaInput) Run() {
	for _, worker := range input.workers {
		go worker.Run()
	}
}

func (input *KafkaInput) Emit() map[string]interface{} {
	return <-input.docCh
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

func (consumer *Consumer) decode(val []byte) map[string]interface{} {
	result := map[string]interface{}{}
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
