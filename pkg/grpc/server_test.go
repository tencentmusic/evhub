package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil)
	assert.Equal(t, defaultServerAddr, s.config.Addr)
	assert.Equal(t, s.server, s.Server())
}

func TestServerServeNilConfig(t *testing.T) {
	s := &Server{}
	err := s.Serve()
	assert.Error(t, err)
}
