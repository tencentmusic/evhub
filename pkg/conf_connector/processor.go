package conf_connector

import (
	"encoding/json"
	"fmt"
)

// ProcessorConfInfo is the configuration information for a processor
type ProcessorConfInfo struct {
	AppID               string              `json:"app_id"`
	TopicID             string              `json:"topic_id"`
	ProtocolType        string              `json:"protocol_type"`
	Addr                string              `json:"addr"`
	DispatcherID        string              `json:"dispatcher_id"`
	IsFlowControl       bool                `json:"is_flow_control"`
	FlowControlStrategy FlowControlStrategy `json:"flow_control_strategy"`
	IsRetry             bool                `json:"is_retry"`
	RetryStrategy       RetryStrategy       `json:"retry_strategy"`
	Timeout             int64               `json:"timeout"`
	IsStop              bool                `json:"is_stop"`
	FilterTag           map[string]bool     `json:"filter_tag"`
	DelayProcessKey     string              `json:"delay_process_key"`
}

// FlowControlStrategy is the strategy of flow control
type FlowControlStrategy struct {
	StrategyType string `json:"strategy_type"`
	QPS          int    `json:"qps"`
}

// RetryStrategy is the strategy of retry
type RetryStrategy struct {
	StrategyType  string `json:"strategy_type"`
	RetryInterval int64  `json:"retry_interval"`
	RetryCount    int    `json:"retry_count"`
}

const (
	ConfigProcessorGroupName = "/evhub/processor/"
)

// KeyProcessorConf returns the key of the processor group
func KeyProcessorConf(dispatcherID string) string {
	return fmt.Sprintf(ConfigProcessorGroupName+"%v", dispatcherID)
}

// ProcessorConfConnector is the interface for a connector of processor configuration
type ProcessorConfConnector interface {
	GetAllConfInfo() ([]*ProcessorConfInfo, error)
	SetWatch(func(c *ProcessorConfInfo, eventType int32) error) error
	SetConfInfo(dispatcherID string, confInfo *ProcessorConfInfo) error
	GetConfInfo(dispatcherID string) (*ProcessorConfInfo, error)
}

// ProcessorConf is processor configuration
type ProcessorConf struct {
	ConfConnector ConfConnector
}

// NewProcessorConfConnector creates a new processor configuration connector
func NewProcessorConfConnector(c ConfConnectorConfig) (ProcessorConfConnector, error) {
	cc, err := NewConfConnector(c, ConfigProcessorGroupName)
	if err != nil {
		return nil, err
	}
	return &ProcessorConf{cc}, nil
}

// GetAllConfInfo returns all configuration information for all topics
func (s *ProcessorConf) GetAllConfInfo() ([]*ProcessorConfInfo, error) {
	rsps, err := s.ConfConnector.GetAllConfInfo()
	if err != nil {
		return nil, err
	}
	confInfoList := make([]*ProcessorConfInfo, 0, len(rsps))
	for _, rsp := range rsps {
		var producerConfInfo ProcessorConfInfo
		cByte, ok := rsp.([]byte)
		if !ok {
			return nil, fmt.Errorf("convert err")
		}
		if err := json.Unmarshal(cByte, &producerConfInfo); err != nil {
			return nil, err
		}
		confInfoList = append(confInfoList, &producerConfInfo)
	}
	return confInfoList, nil
}

// SetWatch is used to watch the configuration
func (s *ProcessorConf) SetWatch(watch func(c *ProcessorConfInfo, eventType int32) error) error {
	err := s.ConfConnector.SetWatch(func(c interface{}, eventType int32) error {
		var producerConfInfo ProcessorConfInfo
		cByte, ok := c.([]byte)
		if !ok {
			return fmt.Errorf("convert err")
		}
		if err := json.Unmarshal(cByte, &producerConfInfo); err != nil {
			return err
		}
		if err := watch(&producerConfInfo, eventType); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// SetConfInfo is used to set the configuration information for a given processor
func (s *ProcessorConf) SetConfInfo(dispatcherID string, confInfo *ProcessorConfInfo) error {
	return s.ConfConnector.SetConfInfo(KeyProcessorConf(dispatcherID), confInfo)
}

// GetConfInfo is used to get the configuration information for a given dispatcherID
func (s *ProcessorConf) GetConfInfo(dispatcherID string) (*ProcessorConfInfo, error) {
	b, err := s.ConfConnector.GetConfInfo(KeyProcessorConf(dispatcherID))
	if err != nil {
		return nil, err
	}
	var r ProcessorConfInfo
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
