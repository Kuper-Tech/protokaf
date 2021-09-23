package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

type ConsumerGroup struct {
	client sarama.ConsumerGroup
}

// NewConsumerGroup creates new ConsumerGroup.
func NewConsumerGroup(brokers []string, group string, config *sarama.Config) (*ConsumerGroup, error) {
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, err
	}

	return &ConsumerGroup{client}, nil
}

// Close closes consumer.
func (c *ConsumerGroup) Close() error {
	return c.client.Close()
}

// Errors returns errors channel.
func (c *ConsumerGroup) Errors() <-chan error {
	return c.client.Errors()
}

// Consume starts non blocking consumer loop for ConsumerHandle on provided topics list.
// Returned channel will closed as soon as Setup step is happened is called for first handler call or if error happens first.
func (c *ConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		if err := c.client.Consume(ctx, topics, handler); err != nil {
			return err
		}
	}
}
