package kafka

import (
	"context"
	"sync"

	"github.com/VertexC/log-formatter/logger"

	"github.com/Shopify/sarama"
)

type worker struct {
	consumer *Consumer
	client   sarama.ConsumerGroup
	topic    string
	logger   *logger.Logger
	docCh    chan map[string]interface{}
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (w *worker) run() {
	w.wg.Add(1)
	topics := []string{w.topic}
	go func() {
		defer w.wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := w.client.Consume(w.ctx, topics, w.consumer); err != nil {
				w.logger.Error.Printf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if w.ctx.Err() != nil {
				return
			}
			// as a hint signal
			w.consumer.ready = make(chan bool)
		}
	}()

	<-w.consumer.ready // Await till the consumer has been set up
	w.logger.Info.Println("Sarama consumer up and running!...")

	select {
	case <-w.ctx.Done():
		w.logger.Info.Println("terminating: context cancelled")
	}
}

func (w *worker) stop() {
	w.cancel()
	<-w.ctx.Done()
	w.wg.Wait()
	if err := w.client.Close(); err != nil {
		w.logger.Warning.Printf("Error closing client: %v", err)
	}
	w.logger.Info.Println("A sarama consumer end!")
}
