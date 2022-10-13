package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryClientInterceptor returns a client timeout interceptor.
func UnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if timeout <= 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}
