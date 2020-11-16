package kafka

import (
	"fmt"
	// "time"
	"github.com/Shopify/sarama"
	"github.com/VertexC/log-formatter/util"
)

var logger = new(util.Logger)

type Config struct {
	Broker    string `yaml:"broker"`
	BatchSize int    `default:"1000" yaml:"batch_size"`
	GroupName string `default:"log-formatter" yaml:"group_name"`
	Topic     string `yaml:"topic"`
	Version   string `default:"2.4.0" yaml:"version"`
	Schema    string `yaml:"schema"`
}

func ExecuteClient(input Config, inputCh chan interface{}, logFile string, verbose bool) {

	logger.Init(logFile, "Kafka-Consumer-Client", verbose)

	config := sarama.NewConfig()
	config.ClientID = input.GroupName
	config.Consumer.Return.Errors = true

	brokers := []string{input.Broker}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topics, _ := master.Topics()

	consumer, errors := consume(input.Topic, topics, master)

	// Count how many message processed
	msgCount := 0

	for {
		select {
		case msg := <-consumer:
			msgCount++
			logger.Trace.Printf("Received messages %+v\n", msg)
			inputCh <- map[string]interface{}{"message": string(msg.Value)}
		case consumerError := <-errors:
			msgCount++
			logger.Error.Fatalln("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
		}
	}
}

func consume(targetTopic string, topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if topic != targetTopic {
			continue
		}
		partitions, _ := master.Partitions(topic)
		for _, partition := range partitions {
			consumer, err := master.ConsumePartition(topic, partition, sarama.OffsetOldest)
			if nil != err {
				fmt.Printf("Topic %v Partitions: %v", topic, partition)
				panic(err)
			}
			fmt.Println(" Start consuming topic ", topic)
			go func(topic string, consumer sarama.PartitionConsumer) {
				for {
					select {
					case consumerError := <-consumer.Errors():
						errors <- consumerError
						logger.Error.Println("consumerError: ", consumerError.Err)

					case msg := <-consumer.Messages():
						consumers <- msg
						logger.Trace.Printf("Got message on topic %s : %s\n", topic, string(msg.Value))
					}
				}
			}(topic, consumer)
		}
	}

	return consumers, errors
}
