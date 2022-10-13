package interceptor

import (
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/tencentmusic/evhub/pkg/grpc/interceptor/statuscode"
	"google.golang.org/grpc"
)

// DefaultUnaryServerInterceptors returns the default unary server interceptors.
func DefaultUnaryServerInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		statuscode.UnaryServerInterceptor(),
	}
}

// DefaultUnaryClientInterceptors returns the default unary client interceptors.
func DefaultUnaryClientInterceptors() []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(opentracing.GlobalTracer())),
	}
}

func DefaultServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(DefaultUnaryServerInterceptors()...),
	}
}
