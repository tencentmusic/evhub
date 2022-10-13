package define

import (
	"fmt"
	"time"

	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
)

const (
	EventTypeInternal = 1
	EventTypeExternal = 2
)

const (
	ProtocolTypeGRPCSend = "grpcSend"
	ProtocolTypeHTTPSend = "httpSend"

	ProtocolTypeGRPCInternalHandler = "grpcInternalHandler"
)

// ConnectorType connector type
type ConnectorType string

const (
	ConnectorTypePulsar ConnectorType = "pulsar"
	ConnectorTypeNsq    ConnectorType = "nsq"

	ConnectorTypeKafka ConnectorType = "kafka"
)

// EventType event type
type EventType int

const (
	RepeatDuplicateScript = `
		local topicRepeatDuplicateKey = KEYS[1]
		local repeatDuplicateTimeOut = ARGV[1]
		local value = redis.call("GET",topicRepeatDuplicateKey)
		if value then
			if (value == "1") then 
				return {"1",value}
			end 
			if (value == "2") then 
				return {"2",value}
			end 
			return {"5",value}
		end
		local setValue = redis.call("SET",topicRepeatDuplicateKey,"1", "EX",repeatDuplicateTimeOut)
		if setValue then
			return {"3",value}
		end
		return {"6",value}
`
)

const (
	RepeatDuplicatePre       = "1"
	RepeatDuplicateCommitted = "2"
	RepeatDuplicatePreFirst  = "3"
	RepeatDuplicateGetErr    = "5"
	RepeatDuplicateSetErr    = "6"
)

const (
	KeyRepeatDuplicate = "evhub:processor:repeat:duplicate:%v"

	RepeatDuplicateTimeOut = 60
)

// TopicRepeatDuplicateKey returns the duplicate key for a topic
func TopicRepeatDuplicateKey(dispatcherEventID string) string {
	return fmt.Sprintf(KeyRepeatDuplicate, dispatcherEventID)
}

// SendReq is req of the dispatcher
type SendReq struct {
	DispatchID  string
	EventID     string
	Event       *comm.Event
	RetryTimes  uint32
	RepeatTimes uint32
	PassBack    []byte
	EndPoint    *EndPoint
}

// EndPoint information about the address
type EndPoint struct {
	Addr    string
	Timeout time.Duration
}

// SendRsp is rsp of the dispatcher
type SendRsp struct {
	PassBack []byte
}

// InternalHandlerReq is req of the internal handler
type InternalHandlerReq struct {
	Msg      *comm.EvhubMsg
	EndPoint *EndPoint
}

// InternalHandlerRsp is rsp of the internal handler
type InternalHandlerRsp struct {
}
