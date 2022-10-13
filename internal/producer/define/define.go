package define

import (
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
)

const (
	ReportEventReport   = "report"
	ReportEventPrepare  = "prepare"
	ReportEventCommit   = "commit"
	ReportEventRollback = "rollback"
	InternalHandler     = "internalHandler"
	ReportEventCallback = "callback"

	PackProtocolHTTPJSON = "httpJson"
	PackProtocolGrpc     = "grpc"
)

// ConnectorType is the name of the connector type
type ConnectorType string

const (
	// ConnectorTypePulsar is the connector type of the pulsar
	ConnectorTypePulsar ConnectorType = "pulsar"
	// ConnectorTypeNsq is the connector type of the nsq
	ConnectorTypeNsq ConnectorType = "nsq"
	// ConnectorTypeKafka is the connector type of the kafka
	ConnectorTypeKafka ConnectorType = "kafka"
)

const (
	EventAppID   string = "event_app_id"
	EventTopicID string = "event_topic_id"
)

// ReportReq is the name of the report request
type ReportReq struct {
	Event   *comm.Event        `json:"event,omitempty"`
	Trigger *comm.EventTrigger `json:"trigger,omitempty"`
	Option  *comm.EventOption  `json:"option,omitempty"`
}

// ReportRsp is the name of the report response
type ReportRsp struct {
	EventID string       `json:"event_id,omitempty"`
	Ret     *comm.Result `json:"ret,omitempty"`
}

// PrepareReq is the name of the prepare request
type PrepareReq struct {
	Event      *comm.Event      `json:"event,omitempty"`
	TxCallback *comm.TxCallback `json:"txCallback,omitempty"`
}

// PrepareRsp is the name of the prepare response
type PrepareRsp struct {
	Tx  *comm.Tx     `json:"tx,omitempty"`
	Ret *comm.Result `json:"ret,omitempty"`
}

// CallbackReq is the name of the callback request
type CallbackReq struct {
	TxEventMsg *TxEventMsg `json:"txEventMsg,omitempty"`
}

// CallbackRsp is the name of the callback response
type CallbackRsp struct {
}

// CommitReq is the name of the commit request
type CommitReq struct {
	EventID string             `json:"eventID,omitempty"`
	Trigger *comm.EventTrigger `json:"trigger,omitempty"`
}

// CommitRsp is the name of the commit response
type CommitRsp struct {
	Ret *comm.Result `json:"ret,omitempty"`
	Tx  *comm.Tx     `json:"tx,omitempty"`
}

// RollbackReq is the name of the rollback request
type RollbackReq struct {
	EventID string `json:"eventID,omitempty"`
}

// RollbackRsp is the name of the rollback response
type RollbackRsp struct {
	Ret *comm.Result `json:"ret,omitempty"`
	Tx  *comm.Tx     `json:"tx,omitempty"`
}

// InternalHandlerReq is the name of the internal handler request
type InternalHandlerReq struct {
	EvhubMsg *comm.EvhubMsg `json:"evhubMsg,omitempty"`
}

// InternalHandlerRsp is the name of the internal handler response
type InternalHandlerRsp struct {
	Ret *comm.Result `json:"ret,omitempty"`
}

// TxEventMsg is the name of the transaction event
type TxEventMsg struct {
	Ctx         interface{}        `json:"ctx"`
	AppID       string             `json:"app_id"`
	TopicID     string             `json:"topic_id"`
	EventID     string             `json:"event_id"`
	DelayTimeMs uint32             `json:"delay_time_ms"`
	RetryCount  uint32             `json:"retry_count"`
	Trigger     *comm.EventTrigger `json:"trigger"`
	Type        int                `json:"type"`
}
