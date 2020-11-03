package kafka

import (
	"fmt"
	"os"
	"os/signal"
	// "time"

	"github.com/Shopify/sarama"
	"github.com/VertexC/log-formatter/util"
)

var logger = new(util.Logger)

type KafkaConfig struct {
	Host      string `yaml:"host"`
	BatchSize int    `default:"1000" yaml:"batch_size"`
	Topic     string `yaml:"topic"`
	Formatter string `yaml:"formatter"`
}

func Execute(input KafkaConfig, inputCh chan interface{}, doneCh chan struct{}) {

	logger.Init("Kafka Consumer")

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true

	brokers := []string{input.Host}

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

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				logger.Trace.Printf("Received messages %+v\n", msg)
				inputCh <- map[string]interface{}{"message": string(msg.Value)}
			case consumerError := <-errors:
				msgCount++
				logger.Error.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				logger.Error.Println("Interrupt is detected")
				doneCh <- struct{}{}
				// default:
				// 	time.Sleep(time.Duration(2) * time.Second)
				// 	logger.Warning.Println("Got nothing!")
			}
		}
	}()
	logger.Info.Println("Processed", msgCount, "messages")
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
