package cmd

import (
	"context"
	"errors"

	"github.com/SberMarket-Tech/protokaf/internal/kafka"
	"github.com/SberMarket-Tech/protokaf/internal/utils/dump"
	"github.com/Shopify/sarama"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConsumeCmd() *cobra.Command {
	var (
		groupFlag  string
		topicsFlag []string
		countFlag  int
		noCommit   bool
	)

	cmd := &cobra.Command{
		Use:   "consume <MessageName>",
		Short: "Consume mode",
		Args:  messageNameRequired,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// parse protofiles & create proto object
			p, err := parseProtofiles()
			if err != nil {
				return
			}

			// find message descriptor
			md, err := findMessage(p, args[0])
			if err != nil {
				return
			}

			if noCommit {
				kafkaConfig.Consumer.Offsets.AutoCommit.Enable = false
			}
			// consumer
			consumer, err := kafka.NewConsumerGroup(viper.GetStringSlice("broker"), groupFlag, kafkaConfig)
			if err != nil {
				return
			}
			defer consumer.Close()

			// start
			go func() {
				log.Infof("Consume topics: %v", topicsFlag)
				if countFlag > 0 {
					log.Infof("Message consuming limit: %d", countFlag)
				}

				handler := &protoHandler{
					MaxCount: countFlag,
					desc:     md,
				}

				err := consumer.Consume(context.Background(), topicsFlag, handler)

				if handler.maximumReached() {
					log.Debugf("Message consuming limit reached: %d", countFlag)
					return
				}

				if err != nil {
					log.Errorf("Consume error: %s", err)
				}

				consumer.Close()
			}()

			// track errors
			for err = range consumer.Errors() {
				if errors.Is(err, ErrMaximumReached) {
					return nil
				}
			}

			return
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&groupFlag, "group", "G", "", "Consumer group")
	flags.StringSliceVarP(&topicsFlag, "topic", "t", []string{}, "Topic to consume from")
	flags.IntVarP(&countFlag, "count", "c", 0, "Exit after consuming this number of messages")
	flags.BoolVar(&noCommit, "no-commit", false, "Consume messages without commiting offset")

	_ = cmd.MarkFlagRequired("group")
	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

type protoHandler struct {
	desc              *desc.MessageDescriptor
	MaxCount, counter int
}

func (protoHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (protoHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ErrMaximumReached error if limit reached
var ErrMaximumReached = errors.New("maximum message reached")

func (h *protoHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	f := dynamic.NewMessageFactoryWithDefaults()

	for msg := range claim.Messages() {
		// skip other partitions
		if p := flags.Partition; p >= 0 && p != msg.Partition {
			continue
		}

		m := f.NewDynamicMessage(h.desc)

		if err := m.Unmarshal(msg.Value); err != nil {
			log.Errorf("Unmarshal message error: %s", err)
		} else {
			dump.DynamicMessage(log, "Message consumed", viper.GetString("output"), m)
		}
		dumpConsumerMessage(msg)

		sess.MarkMessage(msg, "")

		h.counter++
		log.Debugf("Message consuming count: %d", h.counter)

		if h.maximumReached() {
			return ErrMaximumReached
		}
	}

	return nil
}

func (h protoHandler) maximumReached() bool {
	if h.MaxCount != 0 {
		return h.counter == h.MaxCount
	}

	return false
}

func dumpConsumerMessage(msg *sarama.ConsumerMessage) {
	headers := kafka.NewRecordHeadersFromPointers(msg.Headers)

	pairs := dump.Pairs{
		{Name: "timestamp", Value: msg.Timestamp},
		{Name: "topic", Value: msg.Topic},
		{Name: "partition", Value: msg.Partition},
		{Name: "offset", Value: msg.Offset},
		{Name: "key", Value: string(msg.Key)},
		{Name: "length", Value: len(msg.Value)},
		{Name: "headers", Value: headers.String()},
		{Name: "value", Value: msg.Value},
	}
	pairs.Dump(log)
}
