package kafka

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/Shopify/sarama"
	"github.com/xdg/scram"
)

type XDGSCRAMClient struct {
	*scram.ClientConversation
	HashGeneratorFcn scram.HashGeneratorFcn
}

// Begin prepares the client for the SCRAM exchange
// with the server with a user name and a password
func (x *XDGSCRAMClient) Begin(user, password, authzID string) error {
	client, err := x.HashGeneratorFcn.NewClient(user, password, authzID)
	if err != nil {
		return err
	}

	x.ClientConversation = client.NewConversation()

	return nil
}

func scramClientForSASLMechanism(mechanism sarama.SASLMechanism) func() sarama.SCRAMClient {
	var hashGen scram.HashGeneratorFcn

	switch mechanism {
	case sarama.SASLTypeSCRAMSHA256:
		hashGen = sha256.New

	case sarama.SASLTypeSCRAMSHA512:
		hashGen = sha512.New
	}

	if hashGen != nil {
		return func() sarama.SCRAMClient {
			return &XDGSCRAMClient{
				HashGeneratorFcn: hashGen,
			}
		}
	}

	return nil
}
