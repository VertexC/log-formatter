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

var logger = new(util.Logger)

func Execute(config KafkaConfig, output chan interface{}, logFile string, verbose bool) {
	logger.Init(logFile, "[Output-Kafka]", verbose)
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

	for record := range output {
		data, err := json.Marshal(record)
		if err != nil {
			logger.Error.Printf("Failed to parse json from %+v with err %s", record, err)
		}

		// publish without goroutene
		publish(string(data), config.Topic, producer)
	}
}

func publish(message string, topic string, producer sarama.SyncProducer) {
	// publish sync
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	p, o, err := producer.SendMessage(msg)
	if err != nil {
		logger.Trace.Println("Error publish: ", err.Error())
	}
	logger.Trace.Printf("Partition: %d Offset: %d", p, o)
}
