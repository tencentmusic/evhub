package event

import (
	"context"
	"time"

	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"

	"github.com/tencentmusic/evhub/internal/processor/define"
	"github.com/tencentmusic/evhub/internal/processor/options"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Handler is a handler for events of connector
type Handler struct {
	// opts is the name of the processor options
	opts *options.Options
	// handler is the name of the handler
	handler *handler.Handler
	// confInfo is the name of the processor configuration
	confInfo *cc.ProcessorConfInfo
}

// Send dispatch downstream service processing
func (s *Handler) Send(ctx context.Context, msg *comm.EvhubMsg) ([]byte, error) {
	// handle events
	iRsp, err := s.handler.Handler(ctx, &define.SendReq{
		EndPoint: &define.EndPoint{
			Timeout: time.Duration(s.confInfo.Timeout) * time.Millisecond,
			Addr:    s.confInfo.Addr,
		},
		EventID:     msg.EventId,
		Event:       msg.Event,
		RetryTimes:  msg.DispatcherInfo.RetryTimes,
		RepeatTimes: msg.TriggerInfo.RepeatTimes,
		PassBack:    msg.DispatcherInfo.PassBack,
		DispatchID:  msg.DispatcherInfo.DispatcherId,
	}, &handler.UnaryServerInfo{
		FullMethod: s.confInfo.ProtocolType,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.SendRsp)
	if !ok {
		return nil, errcode.CodeError(errcode.CommParamInvalid)
	}
	return rsp.PassBack, nil
}

// InternalHandlerSuccess internal event handling
func (s *Handler) InternalHandlerSuccess(ctx context.Context, msg *comm.EvhubMsg) {
	for {
		if err := s.InternalHandler(ctx, msg); err != nil {
			log.With(ctx).Errorf("send internal handler err: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
}

// InternalHandler internal event handling
func (s *Handler) InternalHandler(ctx context.Context, msg *comm.EvhubMsg) error {
	iRsp, err := s.handler.Handler(ctx, &define.InternalHandlerReq{
		EndPoint: &define.EndPoint{
			Timeout: s.opts.ProducerConfig.Timeout,
			Addr:    s.opts.ProducerConfig.Addr,
		},
		Msg: msg,
	}, &handler.UnaryServerInfo{
		FullMethod: define.ProtocolTypeGRPCInternalHandler,
	})
	if err != nil {
		return err
	}
	_, ok := iRsp.(*define.InternalHandlerRsp)
	if !ok {
		return errcode.CodeError(errcode.CommParamInvalid)
	}
	return nil
}
