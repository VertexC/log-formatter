package kafka

import (
	"github.com/Shopify/sarama"
	"os/signal"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Default *log.Logger
)

type KafkaConfig struct {
	Host string `yaml:"host"`
	BatchSize int `default:"1000" yaml:"batch_size"`
	Topic string `yaml:"topic"`
	Formatter string `yaml:"formatter"`
}

func Init() {
	file, err := os.OpenFile(path.Join("logs", "runtime.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT TRACE]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT INFO]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(io.MultiWriter(file, os.Stdout),
		"[INPUT WARNING]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"[INPUT ERROR]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(os.Stdout,
		"[INPUT DEBUG]: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Default = log.New(io.MultiWriter(file, os.Stdout), "", 0)
}

func Execute(input KafkaConfig, recordCh chan []interface{}, doneCh chan struct{}) {

	Init()

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
				Trace.Printf("Received messages %+v\n", msg)
				record := [] interface{}{
					map[string]interface{}{"message": string(msg.Value)},
				}
				recordCh <- record
			case consumerError := <-errors:
				msgCount++
				Error.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				Error.Println("Interrupt is detected")
				doneCh <- struct{}{}
			default:
				time.Sleep(time.Duration(2)*time.Second)
				Warning.Println("Got nothing!")
			}
		}
	}()
	Info.Println("Processed", msgCount, "messages")
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
						Error.Println("consumerError: ", consumerError.Err)

					case msg := <-consumer.Messages():
						consumers <- msg
						Trace.Printf("Got message on topic %s : %s\n", topic, string(msg.Value))
					}
				}
			}(topic, consumer)
		}
	}

	return consumers, errors
}