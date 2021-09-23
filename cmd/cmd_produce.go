package cmd

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/SberMarket-Tech/protokaf/internal/calldata"
	"github.com/SberMarket-Tech/protokaf/internal/kafka"
	"github.com/SberMarket-Tech/protokaf/internal/proto"
	"github.com/SberMarket-Tech/protokaf/internal/tracing"
	"github.com/SberMarket-Tech/protokaf/internal/utils/dump"
	"github.com/Shopify/sarama"
	"github.com/jhump/protoreflect/desc"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerZapLog "github.com/uber/jaeger-client-go/log/zap"
)

func NewProduceCmd() *cobra.Command { //nolint:funlen,gocognit
	var (
		keyFlag                string
		dataFlag               string
		topicFlag              string
		timeoutStr             string
		timeoutFlag            time.Duration
		traceFlag              bool
		printJaegerConfig      bool
		printTemplateFunctions bool
		countFlag              int
		concurrencyFlag        int
		seedFlag               int64
		headers                []string
	)

	printInfo := func() bool {
		return printTemplateFunctions || printJaegerConfig
	}

	cmd := &cobra.Command{
		Use:   "produce <MessageName>",
		Short: "Produce mode",
		Args: func(cmd *cobra.Command, args []string) (err error) {
			if printInfo() {
				return
			}

			return messageNameRequired(cmd, args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if printInfo() {
				return
			}

			err = cmd.MarkFlagRequired("topic")
			if err != nil {
				return
			}

			if timeoutStr != "" {
				timeoutFlag, err = time.ParseDuration(timeoutStr)
				if err != nil {
					return
				}
			}

			if countFlag < 1 {
				countFlag = 1
			}

			if concurrencyFlag < 1 {
				concurrencyFlag = 1
			}

			return
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if printTemplateFunctions {
				calldata.PrintFuncs(cmd.OutOrStdout())
				return
			}

			// create tracer config
			jaegerCfg, jaegerCfgErr := tracing.NewJaegerConfig()
			if printJaegerConfig {
				dump.PrintStruct(log, "Jaeger config", jaegerCfg)
				return
			}

			// create tracer
			if traceFlag {
				if jaegerCfgErr != nil {
					return jaegerCfgErr
				}

				jaegerLogger := zapLog.Named("jaeger")
				tracer, closer, err := jaegerCfg.NewTracer(
					jaegerConfig.Logger(jaegerZapLog.NewLogger(jaegerLogger)),
				)
				if err != nil {
					return err
				}

				opentracing.SetGlobalTracer(tracer)
				log.Debug("Create new tracer")
				defer closer.Close()
			}

			// parse protofiles & create proto object
			p, err := parseProtofiles()
			if err != nil {
				return
			}

			// find message descriptor
			md, err := findMessage(p, args[0])
			if err != nil {
				return
			}

			// read data form stdin or -d flag
			data, err := readData(dataFlag)
			if err != nil {
				return
			}

			// set seed for random data
			calldata.SetSeeder(seedFlag)

			if countFlag > 1 {
				log.Infof("Producing %d messages...", countFlag)
			}

			// partition num defined by message field
			if p := flags.Partition; p != -1 {
				kafkaConfig.Producer.Partitioner = func(topic string) sarama.Partitioner {
					return constPartitioner{p}
				}
			}

			// create producer
			producer, err := kafka.NewProducer(viper.GetStringSlice("broker"), kafkaConfig)
			if err != nil {
				return err
			}
			defer producer.Close()

			// parse template for data
			tmpl, err := calldata.ParseTemplate(data)
			if err != nil {
				return
			}

			// send messages
			execCtx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			workers := newProduceWorker(concurrencyFlag)
			workers.Run(execCtx, countFlag)

			go func() {
				for i := 0; i < countFlag; i++ {
					workers.AddJob(&produceMessage{
						reqNum:       i,
						key:          keyFlag,
						topic:        topicFlag,
						data:         data,
						headers:      headers,
						sendTimeout:  timeoutFlag,
						producer:     producer,
						traceEnabled: traceFlag,
						tracer:       opentracing.GlobalTracer(),
						tmpl:         tmpl,
						messageDesc:  md,
					})
				}
			}()

			return workers.Result()
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&keyFlag, "key", "k", "", "Message key")
	flags.StringVarP(&dataFlag, "data", "d", "", "Message data")
	flags.StringVarP(&topicFlag, "topic", "t", "", "Topic to produce to")
	flags.StringArrayVarP(&headers, "header", "H", []string{}, "Add message headers (may be specified multiple times)")
	flags.StringVar(&timeoutStr, "timeout", "60s", "Operation timeout")
	flags.BoolVar(&traceFlag, "trace", false, "Send OpenTracing spans to Jaeger")
	flags.BoolVar(&printJaegerConfig, "jaeger-config-print", false, "Print Jaeger config")
	flags.IntVarP(&countFlag, "count", "c", 1, "Producing this number of messages")
	flags.IntVar(&concurrencyFlag, "concurrency", 1, "Number of message senders to run concurrently for const concurrency producing")
	flags.Int64Var(&seedFlag, "seed", 0, "Set seed for pseudo-random sequence")
	flags.BoolVar(&printTemplateFunctions, "template-functions-print", false, "Print functions for using in template")

	tracing.SetJaegerFlags(flags)

	return cmd
}

func makeProduceHeaders(headers []string) []sarama.RecordHeader {
	sep := []byte{'='}
	hdrs := make([]sarama.RecordHeader, 0, len(headers))
	for _, h := range headers {
		parts := bytes.SplitN([]byte(h), sep, 2)
		if len(parts) != 2 {
			log.Warnf("Invalid headers pair: %s", h)
			continue
		}

		hdrs = append(hdrs, sarama.RecordHeader{
			Key:   parts[0],
			Value: parts[1],
		})
	}

	return hdrs
}

func getProducedMessageData(msg *sarama.ProducerMessage) dump.Pairs {
	var keyBytes []byte
	if msg.Key != nil {
		keyBytes, _ = msg.Key.Encode()
	}

	var valuesBytes []byte
	valuesBytes, _ = msg.Value.Encode()

	headers := kafka.RecordHeaders(msg.Headers)

	return dump.Pairs{
		{Name: "topic", Value: msg.Topic},
		{Name: "partition", Value: msg.Partition},
		{Name: "offset", Value: msg.Offset},
		{Name: "key", Value: string(keyBytes)},
		{Name: "length", Value: msg.Value.Length()},
		{Name: "headers", Value: headers.String()},
		{Name: "metadata", Value: msg.Metadata},
		{Name: "value", Value: valuesBytes},
	}
}

func readData(dataFlag string) ([]byte, error) {
	if dataFlag != "" {
		log.Debugf("Read data from --data value: %s", dataFlag)

		return []byte(dataFlag), nil
	}

	// read from stdin
	log.Debug("Reading data from stdin...")

	reader := bufio.NewReader(os.Stdin)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	log.Debugf("Read data from stdin: %s", strings.TrimSpace(string(data)))

	return data, nil
}

type produceMessage struct {
	reqNum       int
	key          string
	topic        string
	data         []byte
	headers      []string
	sendTimeout  time.Duration
	producer     *kafka.Producer
	traceEnabled bool
	tracer       opentracing.Tracer
	tmpl         *template.Template
	messageDesc  *desc.MessageDescriptor
}

func (p *produceMessage) Send(parentCtx context.Context) error {
	cd := calldata.NewCallData(p.reqNum)
	b, err := cd.Execute(p.tmpl)
	if err != nil {
		return err
	}

	// parse data and create message
	m, err := proto.Unmarshal(b.Bytes(), p.messageDesc)
	if err != nil {
		return err
	}
	log.Debugf("Prepared protobuf message: %v", m)

	// message to send
	msg := &sarama.ProducerMessage{
		Topic:   p.topic,
		Key:     sarama.StringEncoder(p.key),
		Value:   proto.Encoder(m),
		Headers: makeProduceHeaders(p.headers),
	}

	ctx, cancel := context.WithTimeout(parentCtx, p.sendTimeout)
	defer cancel()

	var span opentracing.Span

	if p.traceEnabled {
		span, err = tracing.CreateSpan(p.tracer, msg)
		if err != nil {
			return err
		}

		log.Debugf("Create new span: %v", span)
		defer span.Finish()
	}

	if err := p.producer.SendMessage(ctx, msg); err != nil {
		if p.traceEnabled {
			ext.LogError(span, err)
		}

		return err
	}

	dump.DynamicMessage(log, "Message produced", viper.GetString("output"), m)
	getProducedMessageData(msg).Dump(log)

	return nil
}

type produceWorker struct {
	executed    int64
	concurrency int
	jobs        chan *produceMessage
	result      chan error
}

func newProduceWorker(concurrency int) *produceWorker {
	return &produceWorker{
		concurrency: concurrency,
		jobs:        make(chan *produceMessage, concurrency),
		result:      make(chan error, 1),
	}
}

func (p *produceWorker) AddJob(pm *produceMessage) {
	p.jobs <- pm
}

func (p *produceWorker) Result() error {
	return <-p.result
}

func (p *produceWorker) Run(parentCtx context.Context, count int) {
	ctx, cancel := context.WithCancel(parentCtx)

	done := func(err error) {
		cancel()
		p.result <- err
	}

	for i := 0; i < p.concurrency; i++ {
		go func() {
			for {
				select {
				case pm := <-p.jobs:
					if err := pm.Send(ctx); err != nil {
						done(err)
						return
					}

					if int(atomic.AddInt64(&p.executed, 1)) == count {
						done(nil)
						return
					}

				case <-ctx.Done():
					done(ctx.Err())
					return
				}
			}
		}()
	}
}

type constPartitioner struct {
	partition int32
}

func (p constPartitioner) Partition(_ *sarama.ProducerMessage, _ int32) (int32, error) {
	return p.partition, nil
}

func (p constPartitioner) RequiresConsistency() bool {
	return true
}
