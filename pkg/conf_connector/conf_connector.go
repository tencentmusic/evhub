package conf_connector

import (
	"errors"
	"time"
)

const (
	EventTypeNone   = int32(0)
	EventTypeUpdate = int32(1)
	EventTypeAdd    = int32(2)
	EventTypeDelete = int32(3)
	EventTypeAll    = int32(4)
)

// ConfigType is the type of the configuration
type ConfigType string

const (
	ConfigTypeEtcd ConfigType = "etcd"
)

// ConfConnector is the interface for connecting configuration platform
type ConfConnector interface {
	GetAllConfInfo() ([]interface{}, error)
	SetWatch(func(c interface{}, eventType int32) error) error
	SetConfInfo(key string, confInfo interface{}) error
	GetConfInfo(key string) ([]byte, error)
}

// ConfConnectorConfig is connector configuration
type ConfConnectorConfig struct {
	ConfType   ConfigType
	EtcdConfig EtcdConfig
}

// EtcdConfig is etcd configuration
type EtcdConfig struct {
	EtcdEndpoints   []string
	EtcdDialTimeout time.Duration
}

// NewConfConnector creates a new connector configuration
func NewConfConnector(c ConfConnectorConfig, configGroupName string) (ConfConnector, error) {
	switch c.ConfType {
	case ConfigTypeEtcd:
		return NewEtcd(c.EtcdConfig, configGroupName)
	default:
		return nil, errors.New("conf type err")
	}
}
