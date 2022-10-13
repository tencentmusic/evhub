package options

import (
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/config"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/monitor"
)

// Options is a admin server config
type Options struct {
	LogConfig           log.Config
	MonitorConfig       monitor.Options
	ConfConnectorConfig cc.ConfConnectorConfig
	HTTPServerConfig    HTTPServerConfig
}

// HTTPServerConfig is the configuration for HTTP server
type HTTPServerConfig struct {
	Addr string
}

// NewOptions creates a new admin configuration
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
