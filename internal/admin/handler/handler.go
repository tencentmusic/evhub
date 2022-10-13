package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tencentmusic/evhub/internal/admin/define"
	"github.com/tencentmusic/evhub/internal/admin/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
)

// Handler is the handler for events
type Handler struct {
	// ProducerConfConnector is name of the producer configuration
	ProducerConfConnector cc.ProducerConfConnector
	// ProcessorConfConnector is name of the processor configuration
	ProcessorConfConnector cc.ProcessorConfConnector
}

// New creates a new handler
func New(opts *options.Options) (*Handler, error) {
	producerConfConnector, err := cc.NewProducerConfConnector(opts.ConfConnectorConfig)
	if err != nil {
		return nil, err
	}
	processorConfConnector, err := cc.NewProcessorConfConnector(opts.ConfConnectorConfig)
	if err != nil {
		return nil, err
	}
	return &Handler{
		ProducerConfConnector:  producerConfConnector,
		ProcessorConfConnector: processorConfConnector,
	}, nil
}

// SetProducer  creates a new producer configuration
func (s *Handler) SetProducer(c *gin.Context) {
	var req = &define.ProducerReq{}
	var rsp = &define.ProducerRsp{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	// parse request
	if err := c.ShouldBindJSON(req); err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.CommParamInvalid,
			Msg:  err.Error(),
		}
		return
	}

	// handle
	err := s.ProducerConfConnector.SetConfInfo(req.ProducerConfInfo.AppID,
		req.ProducerConfInfo.TopicID, req.ProducerConfInfo)

	if err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  err.Error(),
		}
	}

	rsp.Ret = &comm.Result{}
}

// SetProcessor creates a new processor configuration
func (s *Handler) SetProcessor(c *gin.Context) {
	var req = &define.ProcessorReq{}
	var rsp = &define.ProcessorRsp{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	// parse request
	if err := c.ShouldBindJSON(req); err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.CommParamInvalid,
			Msg:  err.Error(),
		}
		return
	}

	// handle
	err := s.ProcessorConfConnector.SetConfInfo(req.ProcessorConfInfo.DispatcherID, req.ProcessorConfInfo)

	if err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  err.Error(),
		}
	}

	rsp.Ret = &comm.Result{}
}

// GetProcessor gets the processor information
func (s *Handler) GetProcessor(c *gin.Context) {
	var rsp = &define.GetProcessor{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	dispatcherID := c.Param("dispatcher_id")

	// handle
	processorConfInfo, err := s.ProcessorConfConnector.GetConfInfo(dispatcherID)

	if err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  err.Error(),
		}
	}

	rsp.ProcessorConfInfo = processorConfInfo
}

// GetProducer get producer information
func (s *Handler) GetProducer(c *gin.Context) {
	var rsp = &define.GetProducer{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	appID := c.Param("app_id")
	topicID := c.Param("topic_id")

	// handle
	producerConfInfo, err := s.ProducerConfConnector.GetConfInfo(appID, topicID)

	if err != nil {
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  err.Error(),
		}
	}

	rsp.ProducerConfInfo = producerConfInfo
}
