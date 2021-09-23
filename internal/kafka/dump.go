package kafka

import (
	"strings"

	"github.com/Shopify/sarama"
)

type RecordHeaders []sarama.RecordHeader

func (p RecordHeaders) String() string {
	result := strings.Builder{}

	result.WriteByte('{')
	for i, h := range p {
		result.Write(h.Key)
		result.WriteByte(':')
		result.WriteByte('"')
		result.Write(h.Value)
		result.WriteByte('"')

		if len(p) != (i + 1) {
			result.WriteString(", ")
		}
	}
	result.WriteByte('}')

	return result.String()
}

// NewRecordHeadersFromPointers converts []*sarama.RecordHeader -> []sarama.RecordHeader.
func NewRecordHeadersFromPointers(headers []*sarama.RecordHeader) RecordHeaders {
	result := make(RecordHeaders, 0, len(headers))

	for _, h := range headers {
		result = append(result, *h)
	}

	return result
}
