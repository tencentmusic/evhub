package statuscode

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor of the GRPC header status code
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		_ = grpc.SetHeader(ctx, metadata.Pairs("status-code",
			strconv.Itoa(int(status.Code(err)))))
		return resp, err
	}
}
