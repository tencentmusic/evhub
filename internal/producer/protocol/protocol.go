package protocol

import (
	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/internal/producer/protocol/grpc"
	"github.com/tencentmusic/evhub/internal/producer/protocol/http"
	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/program"
)

// Protocol is the protocol of processor events
type Protocol struct {
	// handler is the name of the handler
	handler *handler.Handler
	// opts is the name of the processor configuration
	opts *options.Options
	// protocols is the list of protocols supported
	protocols []program.Svc
}

// New creates a new protocol
func New(h *handler.Handler, opts *options.Options) (*Protocol, error) {
	p := &Protocol{
		handler:   h,
		opts:      opts,
		protocols: make([]program.Svc, 0),
	}

	p.registers(&grpc.Grpc{H: h, Opts: opts})
	p.registers(&http.HTTP{H: h, Opts: opts})

	return p, nil
}

// registers is used to register protocol handlers
func (s *Protocol) registers(p ...program.Svc) {
	s.protocols = append(s.protocols, p...)
}

// Start initialize some operations
func (s *Protocol) Start() error {
	for i := range s.protocols {
		if err := s.protocols[i].Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the protocol handlers gracefully.
func (s *Protocol) Stop() error {
	for _, p := range s.protocols {
		if err := p.Stop(); err != nil {
			log.Errorf("stop protocol err: %v", err)
		}
	}
	return nil
}
