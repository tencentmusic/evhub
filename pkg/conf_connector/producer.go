package conf_connector

import (
	"encoding/json"
	"fmt"
)

// ProducerConfInfo is the configuration information for a producer
type ProducerConfInfo struct {
	AppID              string `json:"app_id"`
	TopicID            string `json:"topic_id"`
	QPS                int64  `json:"qps"`
	MsgSize            int64  `json:"msg_size"`
	TotalMsg           int64  `json:"total_msg"`
	StoragePeriod      int64  `json:"storage_period"`
	TxProtocolType     int    `json:"tx_protocol_type"`
	TxAddress          string `json:"tx_address"`
	TxTimeout          int64  `json:"tx_timeout"`
	TxCallbackInterval int64  `json:"tx_callback_interval"`
}

const (
	ConfigProducerGroupName = "/evhub/producer/"
)

// KeyProducerConf returns the key of producer group
func KeyProducerConf(appID, topicID string) string {
	return fmt.Sprintf(ConfigProducerGroupName+"%v_%v", appID, topicID)
}

// ProducerConfConnector is the interface for a connector of producer configuration
type ProducerConfConnector interface {
	GetAllConfInfo() ([]*ProducerConfInfo, error)
	SetWatch(func(c *ProducerConfInfo, eventType int32) error) error
	SetConfInfo(appID, topicID string, confInfo *ProducerConfInfo) error
	GetConfInfo(appID, topicID string) (*ProducerConfInfo, error)
}

// ProducerConf is producer configuration
type ProducerConf struct {
	// ConfConnector is the connector of producer configuration
	ConfConnector ConfConnector
}

// NewProducerConfConnector creates a new producer configuration connector
func NewProducerConfConnector(c ConfConnectorConfig) (ProducerConfConnector, error) {
	cc, err := NewConfConnector(c, ConfigProducerGroupName)
	if err != nil {
		return nil, err
	}
	return &ProducerConf{cc}, nil
}

// GetAllConfInfo returns all configuration information for all topics
func (s *ProducerConf) GetAllConfInfo() ([]*ProducerConfInfo, error) {
	rsps, err := s.ConfConnector.GetAllConfInfo()
	if err != nil {
		return nil, err
	}
	confInfoList := make([]*ProducerConfInfo, 0, len(rsps))
	for _, rsp := range rsps {
		var producerConfInfo ProducerConfInfo
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
func (s *ProducerConf) SetWatch(watch func(c *ProducerConfInfo, eventType int32) error) error {
	err := s.ConfConnector.SetWatch(func(c interface{}, eventType int32) error {
		var producerConfInfo ProducerConfInfo
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

// SetConfInfo is used to set the configuration information for a given producer
func (s *ProducerConf) SetConfInfo(appID, topicID string, confInfo *ProducerConfInfo) error {
	return s.ConfConnector.SetConfInfo(KeyProducerConf(appID, topicID), confInfo)
}

// GetConfInfo is used to get the configuration information for a given appID, topicID
func (s *ProducerConf) GetConfInfo(appID, topicID string) (*ProducerConfInfo, error) {
	b, err := s.ConfConnector.GetConfInfo(KeyProducerConf(appID, topicID))
	if err != nil {
		return nil, err
	}
	var r ProducerConfInfo
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
