package tracing

import (
	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	tracingProduceOperationName = "produce"
	tracingMessageTopicTag      = "kafka.message.topic"
	tracingMessageLengthTag     = "kafka.message.length"
)

func CreateSpan(tracer opentracing.Tracer, msg *sarama.ProducerMessage) (opentracing.Span, error) {
	if tracer == nil {
		tracer = opentracing.GlobalTracer()
	}

	msgValueLen := 0
	if v := msg.Value; v != nil {
		msgValueLen = v.Length()
	}

	tags := opentracing.Tags{
		tracingMessageTopicTag:  msg.Topic,
		tracingMessageLengthTag: msgValueLen,
	}
	span := tracer.StartSpan(
		tracingProduceOperationName,
		tags,
	)
	ext.SpanKindProducer.Set(span)

	// get OT header
	headers := opentracing.TextMapCarrier{}
	err := tracer.Inject(span.Context(), opentracing.TextMap, &headers)
	if err != nil {
		span.Finish()
		return nil, err
	}

	// set message headers
	msgHeaders := make([]sarama.RecordHeader, 0, len(headers))
	for k, v := range headers {
		msgHeaders = append(msgHeaders, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	msg.Headers = append(msg.Headers, msgHeaders...)

	return span, nil
}
