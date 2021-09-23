package tracing

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestCreateSpan(t *testing.T) {
	const topic = "test-topic"

	msg := sarama.ProducerMessage{
		Topic: topic,
	}

	tracer := mocktracer.New()
	defer func() {
		finished := tracer.FinishedSpans()
		span := finished[0]

		keys := []string{}
		for _, h := range msg.Headers {
			keys = append(keys, string(h.Key))
		}

		assert.Contains(t, keys, "mockpfx-ids-traceid")
		assert.Equal(t, topic, span.Tag(tracingMessageTopicTag))
		assert.Equal(t, ext.SpanKindProducer.Value, span.Tag(string(ext.SpanKind)))
	}()

	span, err := CreateSpan(tracer, &msg)
	if err != nil {
		t.Fatal(err)
	}
	defer span.Finish()
}
