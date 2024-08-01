package cmd

import (
	"fmt"
	"strings"

	"github.com/kuper-tech/protokaf/internal/kafka"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DecodeFlagTextValue is a value of output in text format.
	DecodeFlagTextValue = "text"

	// DecodeFlagJSONValue is a value of output in json format.
	DecodeFlagJSONValue = "json"
)

var decodeFlagValidValues = []string{
	DecodeFlagTextValue,
	DecodeFlagJSONValue,
}

type Flags struct {
	Config    string
	Partition int32

	proto        []string
	broker       []string
	kafkaAuthDSN string
	debug        bool
	output       string

	parent *cobra.Command
}

func NewFlags(parent *cobra.Command) *Flags {
	return &Flags{parent: parent}
}

func (f *Flags) Init() {
	pf := f.parent.PersistentFlags()

	// app flags
	pf.BoolVar(&f.debug, "debug", false, "Enable debugging")

	// kafka
	pf.StringSliceVarP(&f.broker, "broker", "b", []string{"0.0.0.0:9092"}, "Bootstrap broker(s) (host[:port],...)")
	pf.Int32VarP(&f.Partition, "partition", "p", -1, "Partition number")
	pf.StringVarP(&f.kafkaAuthDSN, "kafka-auth-dsn", "X", "", fmt.Sprintf("Kafka auth DSN (%s)", kafka.AuthDSNTemplate))

	// proto
	pf.StringSliceVarP(&f.proto, "proto", "f", []string{}, "Proto files ({file | pattern | url},...)")
	pf.StringVar(&f.output, "output", "json", fmt.Sprintf("Output type: %s", strings.Join(decodeFlagValidValues, ", ")))

	// config
	pf.StringVarP(&f.Config, "config", "F", "", "Config file (default is $HOME/.protokaf.yaml)")

	for _, name := range []string{
		"proto",
		"debug",
		"broker",
		"output",
		"kafka-auth-dsn",
	} {
		_ = viper.BindPFlag(name, pf.Lookup(name))
	}
}

func (f *Flags) Prepare() (err error) {
	found := false
	for _, v := range decodeFlagValidValues {
		if v == f.output {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf(
			"decode flag has invalid value: %s, use one of %s",
			f.output,
			strings.Join(decodeFlagValidValues, ", "),
		)
	}

	return
}
