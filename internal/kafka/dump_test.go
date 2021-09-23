package kafka

import (
	"testing"

	"github.com/Shopify/sarama"
)

func TestRecordHeaders_String(t *testing.T) {
	tests := []struct {
		name string
		p    RecordHeaders
		want string
	}{
		{"empty", RecordHeaders{}, `{}`},
		{"records", RecordHeaders{
			sarama.RecordHeader{
				Key:   []byte("abc"),
				Value: []byte("data"),
			},
		}, `{abc:"data"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("RecordHeaders.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
