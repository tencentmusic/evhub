package protocol

import (
	"github.com/tencentmusic/evhub/internal/processor/define"
	"github.com/tencentmusic/evhub/internal/processor/interceptor"
	"github.com/tencentmusic/evhub/internal/processor/options"
	"github.com/tencentmusic/evhub/internal/processor/protocol/grpc"
	"github.com/tencentmusic/evhub/internal/processor/protocol/http"
	"github.com/tencentmusic/evhub/pkg/handler"
)

// Protocol is the protocol of producer events
type Protocol struct {
	// Handler is the name of the handler
	Handler *handler.Handler
	// opts is the name of the processor options
	opts *options.Options
}

// New creates a new protocol
func New(opts *options.Options) (*Protocol, error) {
	grpcProtocol := &grpc.Grpc{Opts: opts}
	httpProtocol := &http.HTTP{
		Opts: opts,
	}
	methodHandler := map[string]handler.UnaryHandler{
		define.ProtocolTypeGRPCSend:            grpcProtocol.Send,
		define.ProtocolTypeHTTPSend:            httpProtocol.Send,
		define.ProtocolTypeGRPCInternalHandler: grpcProtocol.InternalHandler,
	}
	// initialize handler with interceptors
	h, err := handler.New(methodHandler, interceptor.Register()...)
	if err != nil {
		return nil, err
	}

	return &Protocol{
		opts:    opts,
		Handler: h,
	}, nil
}

// Stop stops the protocol gracefully.
func (s *Protocol) Stop() error {
	return nil
}
