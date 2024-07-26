package cmd

import (
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/kuper-tech/protokaf/internal/kafka"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	flags       *Flags
	kafkaConfig *sarama.Config
)

func Execute() {
	defer func(err error) {
		if err != nil {
			log.Errorf("Error: %s", err)
		}
		_ = log.Sync()

		if err != nil {
			os.Exit(1)
		}
	}(NewRootCmd().Execute())
}

func init() {
	cobra.EnableCommandSorting = false

	setLogger(os.Stdout, "info", appName)
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     appName,
		Version: "1.0.0",
		Short:   fmt.Sprintf("%s - Kafka producer and consumer tool in protobuf format", appName),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			configFiles, err := initConfig()
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return
				}
			}

			err = flags.Prepare()
			if err != nil {
				return
			}

			if viper.GetBool("debug") {
				setLogger(cmd.OutOrStdout(), "debug", cmd.CalledAs())
				log.Info("Debugging enabled")

				sarama.Logger = zap.NewStdLog(zapLog.Named("kafka"))
			}

			if configFiles != "" {
				log.Debugf("Using config file: %s", configFiles)
			}

			kafkaConfig, err = kafka.NewConfig(appName, viper.GetString("kafka-auth-dsn"))
			if err != nil {
				return
			}

			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags = NewFlags(cmd)
	flags.Init()

	cmd.AddCommand(
		NewProduceCmd(),
		NewConsumeCmd(),
		NewListCmd(),
		NewBuildCmd(),
	)

	return cmd
}
