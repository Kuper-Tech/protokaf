package cmd

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewListCmd() *cobra.Command {
	var (
		topicsFlag []string
	)

	cmd := &cobra.Command{
		Use:   "list <MessageName>",
		Short: "Metadata listing",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// client
			client, err := sarama.NewClient(viper.GetStringSlice("broker"), kafkaConfig)
			if err != nil {
				return
			}
			defer client.Close()

			// get topics
			topics, err := client.Topics()
			if err != nil {
				return fmt.Errorf("kafka client got error: %s", err)
			}

			brokers := client.Brokers()

			// find connected broker
			var broker *sarama.Broker
			for _, b := range brokers {
				if err := b.Open(kafkaConfig); err == nil {
					broker = b
					break
				}
			}
			if broker == nil {
				return fmt.Errorf("failed to connect to any of the given brokers (%v) for metadata request", brokers)
			}

			// get metadata from broker
			metadata, err := broker.GetMetadata(&sarama.MetadataRequest{
				Topics: topics,
			})
			if err != nil {
				return fmt.Errorf("get metadata got error: %s", err)
			}

			// output
			log.Infof("%d brokers:", len(metadata.Brokers))
			for _, b := range metadata.Brokers {
				log.Infof(` broker %d "%s"`, b.ID(), b.Addr())
			}

			list := filteredTopics(metadata.Topics, topicsFlag)

			log.Infof("%d topics:", len(list))
			for _, t := range list {
				partitions := t.Partitions

				log.Infof(`  topic "%s", partitions: %d`, t.Name, len(partitions))
				for _, p := range partitions {
					log.Infof(
						`    partition %d, leader %d, replicas: %d (offline: %d), isrs: %d`,
						p.ID, p.Leader, p.Replicas, p.OfflineReplicas, p.Isr,
					)
				}
			}

			return
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVarP(&topicsFlag, "topic", "t", []string{}, "Topic(s) to query (optional)")

	return cmd
}

func filteredTopics(mdTopics []*sarama.TopicMetadata, topics []string) (result []*sarama.TopicMetadata) {
	if len(topics) == 0 {
		return mdTopics
	}

	for _, mt := range mdTopics {
		for _, lt := range topics {
			if mt.Name == lt {
				result = append(result, mt)
				break
			}
		}
	}

	return
}
