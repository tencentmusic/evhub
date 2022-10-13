package grpc

import (
	"errors"
	"time"

	"github.com/tencentmusic/evhub/pkg/grpc/interceptor/timeout"

	"github.com/tencentmusic/evhub/pkg/grpc/interceptor"
	"github.com/tencentmusic/evhub/pkg/grpc/pool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultClientTimeout          = 1000 * time.Millisecond
	defaultClientKeepAliveTime    = 10 * time.Second
	defaultClientKeepAliveTimeout = 3 * time.Second
)

// ClientConfig is a gRPC client config.
type ClientConfig struct {
	// Addr is the gRPC server listen address.
	Addr string
	// RPC timeout.
	Timeout time.Duration
	// KeepAliveTime is the gRPC server keep alive timeout
	KeepAliveTime time.Duration
	// KeepAliveTimeout is the gRPC server keep alive timeout
	KeepAliveTimeout time.Duration
	// DisableReuseConn is the gRPC server disable reuse connection
	DisableReuseConn bool
}

// Dial creates a client connection to the given target.
func Dial(config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if config == nil {
		return nil, errors.New("grpc: empty config")
	}
	if config.Addr == "" {
		return nil, errors.New("grpc: empty address")
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultClientTimeout
	}
	if config.KeepAliveTime <= 0 {
		config.KeepAliveTime = defaultClientKeepAliveTime
	}
	if config.KeepAliveTimeout <= 0 {
		config.KeepAliveTimeout = defaultClientKeepAliveTimeout
	}

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithChainUnaryInterceptor(timeout.UnaryClientInterceptor(config.Timeout)))
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    config.KeepAliveTime,
		Timeout: config.KeepAliveTimeout,
	}))
	return grpc.Dial(config.Addr, opts...)
}

// NewPool creates a new pool
func NewPool(addr string, init, capacity int, idleTimeout time.Duration, timeout time.Duration) (*pool.Pool, error) {
	p, err := pool.New(func() (*grpc.ClientConn, error) {
		return Dial(&ClientConfig{
			Addr:    addr,
			Timeout: timeout,
		}, grpc.WithChainUnaryInterceptor(interceptor.DefaultUnaryClientInterceptors()...))
	}, init, capacity, idleTimeout)
	if err != nil {
		return nil, err
	}
	return p, nil
}
