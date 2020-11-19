package kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/VertexC/log-formatter/util"
)

type KafkaConfig struct {
	Broker string `yaml:"broker"`
	Topic  string `yaml:"topic"`
}

type KafkaOutput struct {
	logger   *util.Logger
	docCh    chan map[string]interface{}
	producer sarama.SyncProducer
	config   KafkaConfig
}

func NewKafkaOutput(config KafkaConfig, docCh chan map[string]interface{}) *KafkaOutput {
	logger := util.NewLogger("[Output-Kafka]")
	sarama.Logger = logger.Trace

	// producer config
	saramCfg := sarama.NewConfig()
	saramCfg.Producer.Retry.Max = 5
	saramCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramCfg.Producer.Return.Successes = true

	// async producer
	//prd, err := sarama.NewAsyncProducer([]string{kafkaConn}, config)

	// sync producer
	producer, err := sarama.NewSyncProducer([]string{config.Broker}, saramCfg)

	if err != nil {
		logger.Error.Fatalln("Error producer: ", err.Error())
	}

	output := &KafkaOutput{
		logger:   logger,
		docCh:    docCh,
		producer: producer,
		config:   config,
	}
	return output
}

func (output *KafkaOutput) Run() {
	logger := output.logger
	for doc := range output.docCh {
		data, err := json.Marshal(doc)
		if err != nil {
			logger.Error.Printf("Failed to parse json from %+v with err %s", doc, err)
		}
		message := string(data)
		// publish wi	thout goroutene
		msg := &sarama.ProducerMessage{
			Topic: output.config.Topic,
			Value: sarama.StringEncoder(message),
		}
		p, o, err := output.producer.SendMessage(msg)
		if err != nil {
			logger.Warning.Println("Error publish: ", err.Error())
		}
		logger.Trace.Printf("Partition: %d Offset: %d", p, o)
	}
}
