package options

import (
	"time"

	"github.com/tencentmusic/evhub/internal/producer/define"

	"github.com/tencentmusic/evhub/pkg/monitor"

	"github.com/tencentmusic/evhub/pkg/bizno"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/config"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Options is a producer server config
type Options struct {
	LogConfig               log.Config
	GRPCServerConfig        GRPCServerConfig
	HTTPServerConfig        HTTPServerConfig
	BizNO                   bizno.Config
	RealTimeConnectorConfig ConnectorConfig
	DelayConnectorConfig    ConnectorConfig
	TxConfig                TransactionConfig
	ConfConnectorConfig     cc.ConfConnectorConfig
	MonitorConfig           monitor.Options
}

// GRPCServerConfig is the configuration for gRPC server
type GRPCServerConfig struct {
	Addr string `default:":9000"`
}

// HTTPServerConfig is the configuration for HTTP server
type HTTPServerConfig struct {
	Addr string `default:":8080"`
}

// TransactionConfig is the configuration for transaction
type TransactionConfig struct {
	ReadDBConfig    *DBConfig
	WriteDBConfig   *DBConfig
	RoutinePoolSize int
	Nonblocking     bool
}

// DBConfig is the configuration for database
type DBConfig struct {
	UserName                     string
	Password                     string
	Addr                         string
	Database                     string
	Charset                      string        `default:"utf8mb4"`
	DefaultStringSize            int           `default:"256"`
	MaxIdleConns                 int           `default:"1"`
	MaxOpenConns                 int           `default:"30"`
	ConnMaxLifetime              time.Duration `default:"1h"`
	EnablePrepareStmt            bool          `default:"true"`
	EnableSkipDefaultTransaction bool          `default:"true"`
}

// ConnectorConfig is the configuration for a Connector
type ConnectorConfig struct {
	ConnectorType define.ConnectorType
	PulsarConfig  PulsarConfig
	NsqConfig     NsqConfig
	KafkaConfig   KafkaConfig
}

// PulsarConfig is the configuration for pulsar clusters
type PulsarConfig struct {
	URL         string
	TopicPrefix string
	Token       string
}

// NsqConfig is the configuration for nsq clusters
type NsqConfig struct {
	NodeList []string
}

// KafkaConfig is the configuration for kafka clusters
type KafkaConfig struct {
	BrokerList []string
}

// NewOptions creates a new producer configuration
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
