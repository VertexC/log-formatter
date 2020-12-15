package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/VertexC/log-formatter/agent/output"
	"github.com/VertexC/log-formatter/config"
	"github.com/VertexC/log-formatter/util"
)

type KafkaConfig struct {
	Base   config.ConfigBase
	Broker string `yaml:"broker"`
	Topic  string `yaml:"topic"`
}

type KafkaOutput struct {
	logger   *util.Logger
	docCh    chan map[string]interface{}
	producer sarama.SyncProducer
	config   *KafkaConfig
	saramCfg *sarama.Config
}

func init() {
	err := output.Register("kafka", NewKafkaOutput)
	if err != nil {
		panic(err)
	}
}

func NewKafkaOutput(content interface{}) (output.Output, error) {
	configMapStr, ok := content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Failed to get mapStr for Kafka Output")
	}
	// default config
	config := &KafkaConfig{
		Base: config.ConfigBase{
			Content:          configMapStr,
			MandantoryFields: []string{"broker", "topic"},
		},
	}
	if err := config.Base.Validate(); err != nil {
		return nil, err
	}

	// FIXME: is there a try to assign with type information
	// like: func tryToAssign(a interface{}, b interface{}) error
	if val, ok := configMapStr["broker"].(string); ok {
		config.Broker = val
	} else {
		fmt.Errorf("Failed to convert <broker> field to <string>")
	}
	if val, ok := configMapStr["topic"].(string); ok {
		config.Broker = val
	} else {
		fmt.Errorf("Failed to convert <topic> field to <string>")
	}

	// set log
	logger := util.NewLogger("Output_Kafka")
	sarama.Logger = logger.Trace

	// producer config
	saramCfg := sarama.NewConfig()
	saramCfg.Producer.Retry.Max = 5
	saramCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramCfg.Producer.Return.Successes = true

	// async producer
	//prd, err := sarama.NewAsyncProducer([]string{kafkaConn}, config)

	output := &KafkaOutput{
		logger:   logger,
		docCh:    make(chan map[string]interface{}, 1000),
		config:   config,
		saramCfg: saramCfg,
	}
	return output, nil
}

func (output *KafkaOutput) Run() {
	logger := output.logger

	// FIXME: if broker is unavailble this will report error and quite
	// sync producer
	producer, err := sarama.NewSyncProducer([]string{output.config.Broker}, output.saramCfg)
	if err != nil {
		logger.Error.Fatalln("Error producer: ", err.Error())
	}

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
		p, o, err := producer.SendMessage(msg)
		if err != nil {
			logger.Warning.Println("Error publish: ", err.Error())
		}
		logger.Trace.Printf("Partition: %d Offset: %d", p, o)
	}
}

func (output *KafkaOutput) Send(doc map[string]interface{}) {
	output.docCh <- doc
}
