package proto

import (
	"github.com/Shopify/sarama"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

var _ sarama.Encoder = &protoEncoder{}

type protoEncoder struct {
	data []byte
	err  error
}

// Encoder returns sarama.Encoder for protobuf.
func Encoder(m *dynamic.Message) sarama.Encoder {
	data, err := m.Marshal()

	return &protoEncoder{
		data: data,
		err:  err,
	}
}

func (s protoEncoder) Encode() ([]byte, error) {
	return s.data, s.err
}

func (s protoEncoder) Length() int {
	return len(s.data)
}

func Unmarshal(b []byte, md *desc.MessageDescriptor) (*dynamic.Message, error) {
	f := dynamic.NewMessageFactoryWithDefaults()
	m := f.NewDynamicMessage(md)

	err := m.UnmarshalJSON(b)
	if err != nil {
		return nil, err
	}

	return m, nil
}
