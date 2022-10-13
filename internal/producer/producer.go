package producer

import (
	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/internal/producer/event"
	"github.com/tencentmusic/evhub/internal/producer/interceptor"
	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/internal/producer/protocol"
	"github.com/tencentmusic/evhub/pkg/handler"
)

// Producer is a server of the event
type Producer struct {
	// protocol is the name of the protocol
	protocol *protocol.Protocol
	// e is the name of the event
	e *event.Event
}

// New creates a new producer
func New(opts *options.Options) (*Producer, error) {
	e, err := event.New(opts)
	if err != nil {
		return nil, err
	}

	methodHandler := map[string]handler.UnaryHandler{
		define.ReportEventReport:   e.Report,
		define.ReportEventPrepare:  e.Prepare,
		define.ReportEventCommit:   e.Commit,
		define.ReportEventRollback: e.Rollback,
		define.InternalHandler:     e.InternalHandler,
	}
	h, err := handler.New(methodHandler, interceptor.Register()...)
	if err != nil {
		return nil, err
	}

	p, err := protocol.New(h, opts)
	if err != nil {
		return nil, err
	}
	return &Producer{
		protocol: p,
		e:        e,
	}, nil
}

// Start initialize some operations
func (s *Producer) Start() error {
	if err := s.protocol.Start(); err != nil {
		return err
	}
	return nil
}

// Stop stops the producer gracefully.
func (s *Producer) Stop() error {
	if err := s.protocol.Stop(); err != nil {
		return err
	}
	if err := s.e.Stop(); err != nil {
		return err
	}
	return nil
}
