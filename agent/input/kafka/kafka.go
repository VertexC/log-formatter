package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/VertexC/log-formatter/agent/config"
	"github.com/VertexC/log-formatter/agent/input"
	"github.com/VertexC/log-formatter/agent/input/protocol"
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

type KafkaInput struct {
	logger  *util.Logger
	workers []*worker
	config  KafkaConfig
	docCh   chan map[string]interface{}
}

func init() {
	input.Register("kafka", NewKafkaInput)
}

func NewKafkaInput(content interface{}) (protocol.Input, error) {
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

		client, err := sarama.NewConsumerGroup(config.Brokers, config.GroupName, saramaCfg)
		if err != nil {
			return nil, fmt.Errorf("Error creating consumer group client: %v", err)
		}

		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		input.workers = append(input.workers,
			&worker{
				consumer: consumer,
				client:   client,
				topic:    config.Topic,
				logger:   input.logger,
				ctx:      ctx,
				cancel:   cancel,
				wg:       wg,
			})
	}
	return input, nil
}

func (input *KafkaInput) Run() {
	for _, worker := range input.workers {
		go worker.run()
	}
}

func (input *KafkaInput) Stop() {
	input.logger.Info.Printf("Stop kafka input")
	for _, worker := range input.workers {
		worker.stop()
	}
}

func (input *KafkaInput) Emit() map[string]interface{} {
	return <-input.docCh
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
