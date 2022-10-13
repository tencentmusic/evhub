package grpc

import (
	"context"
	"fmt"

	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/util/addrutil"

	"github.com/tencentmusic/evhub/internal/processor/define"
	"github.com/tencentmusic/evhub/internal/processor/options"
	eh_in "github.com/tencentmusic/evhub/pkg/gen/proto/internal_handler"
	eh_pc "github.com/tencentmusic/evhub/pkg/gen/proto/processor"
	"github.com/tencentmusic/evhub/pkg/grpc"
)

// Grpc is a type of protocol
type Grpc struct {
	// Opts is the name of the processor configuration
	Opts *options.Options
}

// Send dispatch downstream service processing
func (s *Grpc) Send(ctx context.Context, req interface{}) (interface{}, error) {
	// convert request
	in, ok := req.(*define.SendReq)
	if !ok {
		return nil, fmt.Errorf("convert request error: %v", in)
	}
	// parse address
	addr := addrutil.ParseAddr(in.EndPoint.Addr)
	conn, err := grpc.Dial(&grpc.ClientConfig{Addr: addr, Timeout: in.EndPoint.Timeout})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// make request
	dispatchReq := s.makeDispatchReq(in)
	// grpc client
	rsp, err := eh_pc.NewEvhubProcessorClient(conn).Dispatch(ctx, dispatchReq)
	if err != nil {
		return nil, err
	}
	if rsp.Ret != nil && rsp.Ret.Code != 0 {
		return rsp, errcode.CodeError(rsp.Ret.Code)
	}
	// convert response
	return &define.SendRsp{
		PassBack: rsp.PassBack,
	}, nil
}

// makeDispatchReq assembly dispatch req
func (s *Grpc) makeDispatchReq(in *define.SendReq) *eh_pc.DispatchReq {
	return &eh_pc.DispatchReq{
		EventId:     in.EventID,
		Event:       in.Event,
		RetryTimes:  in.RetryTimes,
		RepeatTimes: in.RepeatTimes,
		PassBack:    in.PassBack,
	}
}

// InternalHandler internal event handling
func (s *Grpc) InternalHandler(ctx context.Context, req interface{}) (interface{}, error) {
	// convert request
	in, ok := req.(*define.InternalHandlerReq)
	if !ok {
		return nil, fmt.Errorf("convert request error: %v", in)
	}
	// parse address
	addr := addrutil.ParseAddr(in.EndPoint.Addr)
	conn, err := grpc.Dial(&grpc.ClientConfig{Addr: addr, Timeout: in.EndPoint.Timeout})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// grpc client
	rsp, err := eh_in.NewEvhubInternalHandlerClient(conn).Fork(ctx, &eh_in.ForkReq{
		EvHubMsg: in.Msg,
	})
	if err != nil {
		return nil, err
	}
	if rsp.Ret != nil && rsp.Ret.Code != 0 {
		return rsp, errcode.CodeError(rsp.Ret.Code)
	}
	// convert response
	return &define.InternalHandlerRsp{}, nil
}
