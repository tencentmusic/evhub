package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/tencentmusic/evhub/pkg/log"
)

// ServerConfig is a http server config.
type ServerConfig struct {
	// Addr is the http server listen address.
	Addr string
}

// Server is a http server.
type Server struct {
	// config is the http server configuration
	config *ServerConfig
	// srv is the http server
	srv *http.Server
}

// NewServer creates a new http server
func NewServer(c *ServerConfig) *Server {
	var s = &Server{
		config: c,
	}
	s.srv = &http.Server{
		Addr:    c.Addr,
		Handler: nil,
	}
	return s
}

// SetHandler is used to set handler for the http server
func (s *Server) SetHandler(h http.Handler) {
	s.srv.Handler = h
}

// Serve is used to serve the http server
func (s *Server) Serve() error {
	if s.config == nil {
		return errors.New("http: empty server config")
	}
	lis, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	go func() {
		if sErr := s.srv.Serve(lis); sErr != nil {
			if sErr == http.ErrServerClosed {
				log.Info("http server closed")
				return
			}
			panic(sErr)
		}
	}()
	return nil
}

// Close is used to close the http server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	return nil
}
