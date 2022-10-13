package options

import (
	"time"

	"github.com/tencentmusic/evhub/pkg/redis"

	"github.com/tencentmusic/evhub/internal/processor/define"

	"github.com/tencentmusic/evhub/pkg/monitor"

	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/config"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Options is a processor server config
type Options struct {
	LogConfig           log.Config
	ConnectorConfig     ConnectorConfig
	ConfConnectorConfig cc.ConfConnectorConfig
	RedisConfig         redis.Config
	ProducerConfig      ProducerConfig
	MonitorConfig       monitor.Options
}

// ProducerConfig is a producer server config
type ProducerConfig struct {
	Addr    string
	Timeout time.Duration
}

// ConnectorConfig is a connector config
type ConnectorConfig struct {
	ConnectorType define.ConnectorType
	KafkaConfig   KafkaConfig
	PulsarConfig  PulsarConfig
	NsqConfig     NsqConfig
}

// KafkaConfig is a Kafka config
type KafkaConfig struct {
	BrokerList                 []string
	OffsetsAutoCommitEnable    bool
	FlowControlBatchNum        int
	Version                    string
	InitOffsetGetOffsetTimeout int64
	InitOffsetGetOffsetPeroid  int64
}

// PulsarConfig is a Pulsar config
type PulsarConfig struct {
	URL         string
	TopicPrefix string
	Token       string
}

// NsqConfig is a Nsq config
type NsqConfig struct {
	NodeList    []string
	LookUpAddrs []string
}

// NewOptions creates a new processor configuration
func NewOptions() (*Options, error) {
	opts := &Options{}
	c, err := config.New()
	if err != nil {
		return nil, err
	}
	if err := c.Load(opts); err != nil {
		return nil, err
	}
	return opts, nil
}
