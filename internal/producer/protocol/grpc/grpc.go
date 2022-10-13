package grpc

import (
	"context"
	"fmt"

	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/util/routine"

	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/pkg/gen/proto/internal_handler"
	"github.com/tencentmusic/evhub/pkg/gen/proto/producer"
	"github.com/tencentmusic/evhub/pkg/grpc"
	"github.com/tencentmusic/evhub/pkg/grpc/interceptor"
	"github.com/tencentmusic/evhub/pkg/handler"
	ggrpc "google.golang.org/grpc"
)

// Grpc is a type of protocol
type Grpc struct {
	// H is the name of the handler
	H *handler.Handler
	// Opts is the name of the producer configuration
	Opts *options.Options
	// server is the name of the grpc server
	server *grpc.Server
}

// Start initialize some operations
func (s *Grpc) Start() error {
	opts := []ggrpc.ServerOption{
		ggrpc.ChainUnaryInterceptor(interceptor.DefaultUnaryServerInterceptors()...),
	}
	server := grpc.NewServer(&grpc.ServerConfig{
		Addr: s.Opts.GRPCServerConfig.Addr,
	}, opts...)

	producer.RegisterEvhubProducerServer(server.Server(), s)
	internal_handler.RegisterEvhubInternalHandlerServer(server.Server(), s)

	s.server = server
	routine.GoWithRecovery(func() {
		if err := s.server.Serve(); err != nil {
			log.Panicf("grpc failed to serve: %v", err)
		}
	})
	return nil
}

// Report is used to report events to the server
func (s *Grpc) Report(ctx context.Context, in *producer.ReportReq) (*producer.ReportRsp, error) {
	iRsp, err := s.H.Handler(ctx, &define.ReportReq{
		Event:   in.Event,
		Trigger: in.Trigger,
		Option:  in.Option,
	}, &handler.UnaryServerInfo{
		FullMethod: define.ReportEventReport,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.ReportRsp)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return &producer.ReportRsp{
		EventId: rsp.EventID,
		Ret:     rsp.Ret,
	}, nil
}

// Prepare is used to prepare a transaction message to the server
func (s *Grpc) Prepare(ctx context.Context, in *producer.PrepareReq) (*producer.PrepareRsp, error) {
	iRsp, err := s.H.Handler(ctx, &define.PrepareReq{
		Event:      in.Event,
		TxCallback: in.TxCallback,
	}, &handler.UnaryServerInfo{
		FullMethod: define.ReportEventPrepare,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.PrepareRsp)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return &producer.PrepareRsp{
		Tx:  rsp.Tx,
		Ret: rsp.Ret,
	}, nil
}

// Commit is used to commit a transaction message to the server
func (s *Grpc) Commit(ctx context.Context, in *producer.CommitReq) (*producer.CommitRsp, error) {
	iRsp, err := s.H.Handler(ctx, &define.CommitReq{
		EventID: in.EventId,
		Trigger: in.Trigger,
	}, &handler.UnaryServerInfo{
		FullMethod: define.ReportEventCommit,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.CommitRsp)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return &producer.CommitRsp{
		Tx:  rsp.Tx,
		Ret: rsp.Ret,
	}, nil
}

// Rollback is used to roll back a transaction message to the server
func (s *Grpc) Rollback(ctx context.Context, in *producer.RollbackReq) (*producer.RollbackRsp, error) {
	iRsp, err := s.H.Handler(ctx, &define.RollbackReq{
		EventID: in.EventId,
	}, &handler.UnaryServerInfo{
		FullMethod: define.ReportEventRollback,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.RollbackRsp)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return &producer.RollbackRsp{
		Tx:  rsp.Tx,
		Ret: rsp.Ret,
	}, nil
}

// Fork is used to handle internal messages
func (s *Grpc) Fork(ctx context.Context, in *internal_handler.ForkReq) (*internal_handler.ForkRsp, error) {
	iRsp, err := s.H.Handler(ctx, &define.InternalHandlerReq{
		EvhubMsg: in.EvHubMsg,
	}, &handler.UnaryServerInfo{
		FullMethod: define.InternalHandler,
	})
	if err != nil {
		return nil, err
	}
	rsp, ok := iRsp.(*define.InternalHandlerRsp)
	if !ok {
		return nil, fmt.Errorf("convert err")
	}
	return &internal_handler.ForkRsp{
		Ret: rsp.Ret,
	}, nil
}

// Stop stops the grpc server gracefully.
func (s *Grpc) Stop() error {
	if err := s.server.Close(); err != nil {
		return err
	}
	return nil
}
