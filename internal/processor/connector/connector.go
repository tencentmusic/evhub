package connector

import (
	"errors"

	"github.com/tencentmusic/evhub/internal/processor/connector/pulsar"
	"github.com/tencentmusic/evhub/internal/processor/define"

	"github.com/tencentmusic/evhub/internal/processor/connector/nsq"

	"github.com/tencentmusic/evhub/internal/processor/options"

	"github.com/tencentmusic/evhub/internal/processor/connector/kafka"
	"github.com/tencentmusic/evhub/pkg/handler"

	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
)

// Connector is a store of events
type Connector interface {
	// ConsumerDelayTopic is the subject of the subscription latency processor
	ConsumerDelayTopic(confInfoList []*cc.ProducerConfInfo)
	// WatchDelayTopic watch the topic of the latency processor
	WatchDelayTopic(c *cc.ProducerConfInfo, eventType int32) error
	// ConsumerProcessorTopic is the topic of the subscription processor
	ConsumerProcessorTopic(confInfoList []*cc.ProcessorConfInfo)
	// WatchProcessorTopic watch the topic of the processor
	WatchProcessorTopic(c *cc.ProcessorConfInfo, eventType int32) error
	// Stop stops the connector gracefully.
	Stop() error
}

// New creates the connector of store
func New(opts *options.Options, c *options.ConnectorConfig, handler *handler.Handler) (Connector, error) {
	switch c.ConnectorType {
	case define.ConnectorTypePulsar:
		return pulsar.New(opts, &c.PulsarConfig, handler)
	case define.ConnectorTypeNsq:
		return nsq.New(opts, &c.NsqConfig, handler)
	case define.ConnectorTypeKafka:
		return kafka.New(opts, &c.KafkaConfig, handler)
	default:
		return nil, errors.New("connector type err")
	}
}
