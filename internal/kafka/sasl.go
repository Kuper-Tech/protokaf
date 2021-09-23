package kafka

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
)

const (
	// AuthDSNTemplate template for auth DSN.
	AuthDSNTemplate = "SASLType:login:password"
)

type SASL struct {
	Mechanism sarama.SASLMechanism
	User      string
	Password  string
}

func NewSASL() *SASL {
	return &SASL{}
}

// Parse parses DSN and fill SASL struct.
func (s *SASL) Parse(dsn string) error {
	parts := strings.SplitN(dsn, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf(`incorrect DSN format: expected "%s"`, AuthDSNTemplate)
	}

	s.Mechanism = sarama.SASLMechanism(parts[0])
	s.User = parts[1]
	s.Password = parts[2]

	return nil
}
