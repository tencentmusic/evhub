package grpc

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/tencentmusic/evhub/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultServerAddr = ":9000"
)

// ServerConfig is a gRPC server config.
type ServerConfig struct {
	// Addr is the gRPC server listen address.
	Addr string
}

// Server is a gRPC server to serve RPC requests.
type Server struct {
	config    *ServerConfig
	server    *grpc.Server
	unaryInts []grpc.UnaryServerInterceptor
}

// NewServer creates a gRPC server which has no service registered and has not
// started to accept requests yet.
func NewServer(config *ServerConfig, opts ...grpc.ServerOption) *Server {
	s := new(Server)
	s.SetConfig(config)
	opts = append(opts,
		grpc.ChainUnaryInterceptor(s.unaryInterceptor,
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(WithLogAndMetrics))))
	s.server = grpc.NewServer(opts...)
	return s
}

func WithLogAndMetrics(r interface{}) error {
	log.Errorf("%+v", r)
	return status.Errorf(codes.Internal, "%v", r)
}

// SetConfig sets the default config.
func (s *Server) SetConfig(config *ServerConfig) {
	if config == nil {
		config = new(ServerConfig)
	}
	if config.Addr == "" {
		config.Addr = defaultServerAddr
	}
	s.config = config
}

// Server return the grpc server.
func (s *Server) Server() *grpc.Server {
	return s.server
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each.
func (s *Server) Serve() error {
	if s.config == nil {
		return errors.New("grpc: empty server config")
	}
	lis, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	reflection.Register(s.server)

	if err := s.server.Serve(lis); err != nil {
		panic(err)
	}
	return nil
}

func (s *Server) Close() error {
	s.server.GracefulStop()
	return nil
}

// wait waits for the signal to shut down the server, then gracefully stops it.
func (s *Server) wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	sig := <-ch
	switch sig {
	case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
		s.server.GracefulStop()
		return
	case syscall.SIGHUP:
		return
	}
}

// UseUnaryInterceptor attaches the unary server interceptors to the gRPC server.
func (s *Server) UseUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInts = append(s.unaryInts, interceptors...)
}

// unaryInterceptor chains the all interceptors in Server.unaryInts.
// The first interceptor will be the outer most,
// while the last interceptor will be the inner most wrapper around the real call.
func (s *Server) unaryInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if len(s.unaryInts) == 0 {
		return handler(ctx, req)
	}
	return s.unaryInts[0](ctx, req, info, getChainUnaryHandler(s.unaryInts, 0, info, handler))
}

// getChainUnaryHandler recursively generate the chained UnaryHandler.
func getChainUnaryHandler(interceptors []grpc.UnaryServerInterceptor, curr int,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	if curr == len(interceptors)-1 {
		return handler
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, handler))
	}
}
