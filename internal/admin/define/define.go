package define

import (
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
)

// ProducerReq is the name of the producer request
type ProducerReq struct {
	ProducerConfInfo *cc.ProducerConfInfo `json:"producer_conf_info"`
}

// ProducerRsp is the name of the producer response
type ProducerRsp struct {
	Ret *comm.Result `json:"ret"`
}

// ProcessorReq is the name of the processor request
type ProcessorReq struct {
	ProcessorConfInfo *cc.ProcessorConfInfo `json:"processor_conf_info"`
}

// ProcessorRsp is the name of the processor response
type ProcessorRsp struct {
	Ret *comm.Result `json:"ret"`
}

// GetProcessor is the name of the processor response
type GetProcessor struct {
	ProcessorConfInfo *cc.ProcessorConfInfo `json:"processor_conf_info"`
	Ret               *comm.Result          `json:"ret"`
}

// GetProducer is the name of the producer response
type GetProducer struct {
	ProducerConfInfo *cc.ProducerConfInfo `json:"producer_conf_info"`
	Ret              *comm.Result         `json:"ret"`
}
