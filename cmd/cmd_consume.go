package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/SberMarket-Tech/protokaf/internal/kafka"
	"github.com/SberMarket-Tech/protokaf/internal/utils/dump"
	"github.com/Shopify/sarama"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"sync"
)

func NewConsumeCmd() *cobra.Command {
	var (
		groupFlag  string
		topicsFlag []string
		countFlag  int
		noCommit   bool
		offsets    []string
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

					topics:    topicsFlag,
					partition: int32(0),
				}

				// set offset
				if offsets != nil {
					offsetsArg, err := parseOffsetsFlag(offsets)
					if err != nil {
						log.Errorf("Failed to parse offset: %s", err)
						return
					}
					handler.offsets = offsetsArg
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
	flags.StringSliceVarP(&offsets, "offsets", "o", nil, "Start consuming from this offsets: "+
		"topic.name1:123,topic.name2:321 (offsets for every topic) or 123 (one global offset for all topics), default: newest)")

	_ = cmd.MarkFlagRequired("group")
	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

const globalOffset = "global"

func parseOffsetsFlag(offsetsFlag []string) (offsets map[string]int64, err error) {
	offsets = make(map[string]int64, len(offsetsFlag))
	for _, offset := range offsetsFlag {
		i, err := strconv.ParseInt(offset, 10, 64)
		if err == nil {
			return map[string]int64{globalOffset: i}, nil
		}

		topicOffsetPair := strings.Split(offset, ":")
		if len(topicOffsetPair) != 2 {
			return nil, fmt.Errorf("invalid offset format: %s, "+
				"expected: topic:123 (topic name and offset value) or 123 (global offset value for all topics)", offset)
		}
		offsetForTopic, err := strconv.ParseInt(topicOffsetPair[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid offset format: %s, "+
				"failed to parse offset value: %s\n"+
				"expected: topic:123 (topic name and offset value) or 123 (global offset value for all topics)\n"+
				"error: %s", topicOffsetPair[1], offset, err)
		}
		offsets[topicOffsetPair[0]] = offsetForTopic
	}

	return offsets, nil
}

type protoHandler struct {
	desc              *desc.MessageDescriptor
	MaxCount, counter int
	topics            []string
	partition         int32
	offsets           map[string]int64
}

var once sync.Once

func (p protoHandler) Setup(sess sarama.ConsumerGroupSession) error {
	once.Do(func() {
		goffst, isGlobalOffsetSet := p.offsets[globalOffset]
		for _, topic := range p.topics {
			if partFromFlags := flags.Partition; partFromFlags > 0 {
				p.partition = partFromFlags
			}

			if isGlobalOffsetSet {
				sess.ResetOffset(topic, p.partition, goffst, "")
				continue
			}

			if p.offsets != nil {
				offset, ok := p.offsets[topic]
				if !ok {
					println("ok2")
					continue
				}
				println("ok3")
				sess.ResetOffset(topic, p.partition, offset, "")
			}
		}
	})
	return nil
}

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
