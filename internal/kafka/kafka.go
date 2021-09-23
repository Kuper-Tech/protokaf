package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
)

func NewConfig(appName string, authDSN string) (*sarama.Config, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0 // 0.11 - min version for support record headers
	config.ClientID = appName
	config.Consumer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	if authDSN != "" {
		sasl := NewSASL()

		if err := sasl.Parse(authDSN); err != nil {
			return nil, err
		}

		if err := setupSASL(config, sasl); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func setupSASL(config *sarama.Config, sasl *SASL) error {
	switch sasl.Mechanism {
	case sarama.SASLTypeSCRAMSHA256, sarama.SASLTypeSCRAMSHA512, sarama.SASLTypePlaintext:
	default:
		return fmt.Errorf("unsupported SASL mechanism type: %s", sasl.Mechanism)
	}

	config.Net.SASL.Enable = true
	config.Net.SASL.User = sasl.User
	config.Net.SASL.Password = sasl.Password
	config.Net.SASL.Mechanism = sasl.Mechanism
	config.Net.SASL.SCRAMClientGeneratorFunc = scramClientForSASLMechanism(sasl.Mechanism)

	return nil
}
