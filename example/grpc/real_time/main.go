package main

import (
	"context"
	"time"

	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/grpc"
	"github.com/tencentmusic/evhub/pkg/grpc/interceptor"
	"github.com/tencentmusic/evhub/pkg/log"

	eh_pc "github.com/tencentmusic/evhub/pkg/gen/proto/processor"
	eh_pd "github.com/tencentmusic/evhub/pkg/gen/proto/producer"
	ggrpc "google.golang.org/grpc"
)

// Real-time event example
func main() {
	// the address of the grpc service
	serverAddr := ":6001"
	// address of the Evhub
	addr := "127.0.0.1:9000"
	// client invocation timeout
	timeout := time.Second * 5
	// establish a connection
	conn, err := grpc.Dial(&grpc.ClientConfig{Addr: addr, Timeout: timeout})
	if err != nil {
		log.Panicf("dial err:%v", err)
	}
	defer conn.Close()
	// grpc client
	rsp, err := eh_pd.NewEvhubProducerClient(conn).Report(context.Background(), &eh_pd.ReportReq{
		Event: &comm.Event{
			AppId:   "eh",
			TopicId: "test",
			Data:    []byte("hello world"),
		},
		Trigger: &comm.EventTrigger{
			TriggerType: comm.EventTriggerType_EVENT_TRIGGER_TYPE_REAL_TIME,
		},
	})
	if err != nil {
		log.Panicf("report err:%v", err)
	}
	if rsp.Ret != nil && rsp.Ret.Code != 0 {
		log.Panicf("report failed rsp:%+v", rsp)
	}
	log.Infof("report success rsp:%+v", rsp)
	// start grpc server
	StartGrpcServer(serverAddr, &Svc{})
}

type Svc struct{}

// Dispatch event distribution sink
func (s *Svc) Dispatch(ctx context.Context, req *eh_pc.DispatchReq) (*eh_pc.DispatchRsp, error) {
	log.Infof("req:%+v", req)
	return &eh_pc.DispatchRsp{}, nil
}

// StartGrpcServer start the grpc service
func StartGrpcServer(serverAddr string, s eh_pc.EvhubProcessorServer) {
	// options
	opts := []ggrpc.ServerOption{
		ggrpc.ChainUnaryInterceptor(interceptor.DefaultUnaryServerInterceptors()...),
	}
	// initialization grpc server
	server := grpc.NewServer(&grpc.ServerConfig{
		Addr: serverAddr,
	}, opts...)
	// register evhub Processor Server
	eh_pc.RegisterEvhubProcessorServer(server.Server(), s)
	if err := server.Serve(); err != nil {
		log.Panicf("grpc failed to serve: %v", err)
	}
}
