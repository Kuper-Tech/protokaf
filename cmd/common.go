package cmd

import (
	"fmt"
	"strings"

	"github.com/SberMarket-Tech/protokaf/internal/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName = "protokaf"
)

func parseProtofiles() (*proto.Proto, error) {
	files := viper.GetStringSlice("proto")
	p, err := proto.NewProto(files)
	if err != nil {
		return nil, err
	}

	log.Debugf(`Proto import paths: "%s"`, strings.Join(p.ImportPaths, ":"))

	if len(files) > 0 {
		log.Debugf("Parsed files: %s", strings.Join(files, ", "))
	}

	return p, nil
}

func findMessage(p *proto.Proto, name string) (*desc.MessageDescriptor, error) {
	m, err := p.FindMessage(name)
	if err != nil {
		return nil, err
	}

	log.Debugf("Using proto message: %s", m.GetFullyQualifiedName())

	return m, nil
}

func messageNameRequired(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("required argument <MessageName> not specified")
	}

	return nil
}
