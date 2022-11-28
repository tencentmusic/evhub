package connector

import (
	"context"
	"errors"
	"time"

	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/internal/producer/options"

	"github.com/tencentmusic/evhub/internal/producer/connector/kafka"
	"github.com/tencentmusic/evhub/internal/producer/connector/nsq"
	"github.com/tencentmusic/evhub/internal/producer/connector/pulsar"
)

// Connector is a store of events
type Connector interface {
	// SendMsg is the stores that send real-time events
	SendMsg(ctx context.Context, topic string, orderingKey string, data []byte) error
	// SendDelayMsg is the stores that send delay events
	SendDelayMsg(ctx context.Context, topic string, orderingKey string, data []byte, delay time.Duration) error
	// Stop stops the connector gracefully.
	Stop() error
}

// New creates the connector of store
func New(c *options.ConnectorConfig) (Connector, error) {
	switch c.ConnectorType {
	case define.ConnectorTypePulsar:
		return pulsar.New(c.PulsarConfig.URL, c.PulsarConfig.TopicPrefix,
			c.PulsarConfig.Token)
	case define.ConnectorTypeNsq:
		return nsq.New(c.NsqConfig.NodeList)
	case define.ConnectorTypeKafka:
		return kafka.New(c.KafkaConfig.BrokerList)
	default:
		return nil, errors.New("connector type err")
	}
}
