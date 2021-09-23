package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

// Producer kafka producer.
type Producer struct {
	client sarama.AsyncProducer
}

// NewProducer creates a new Producer using the given broker addresses and configuration.
func NewProducer(brokers []string, config *sarama.Config) (*Producer, error) {
	client, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{client}, nil
}

// SendMessage produces a given message, and returns only when it either has
// succeeded or failed to produce. It will return the partition and the offset
// of the produced message, or an error if the message failed to produce.
func (p *Producer) SendMessage(ctx context.Context, msg *sarama.ProducerMessage) error {
	errs := make(chan error)

	go func() {
		select {
		case p.client.Input() <- msg:
		case <-ctx.Done():
			errs <- ctx.Err()
		}
	}()

	go func() {
		select {
		case err := <-p.client.Errors():
			errs <- err

		case m := <-p.client.Successes():
			*msg = *m
			errs <- nil
		}
	}()

	return <-errs
}

// Close shuts down the producer and waits for any buffered messages to be
// flushed. You must call this function before a producer object passes out of
// scope, as it may otherwise leak memory. You must call this before calling
// Close on the underlying client.
func (p *Producer) Close() error {
	return p.client.Close()
}
