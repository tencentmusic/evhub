package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/tencentmusic/evhub/internal/admin/handler"
	"github.com/tencentmusic/evhub/internal/admin/options"
	eh_http "github.com/tencentmusic/evhub/pkg/http"
)

// Admin is a server of the event
type Admin struct {
	Opts    *options.Options
	Handler *handler.Handler
	server  *eh_http.Server
}

// initialize
func init() {
	gin.SetMode(gin.ReleaseMode)
}

// New creates a new admin server
func New(opts *options.Options) (*Admin, error) {
	h, err := handler.New(opts)
	if err != nil {
		return nil, err
	}

	return &Admin{
		Opts:    opts,
		Handler: h,
	}, nil
}

// Start initialize some operations
func (s *Admin) Start() error {
	var r = gin.New()
	r.POST("v1/producer", s.Handler.SetProducer)
	r.POST("v1/processor", s.Handler.SetProcessor)
	r.GET("v1/producer/:app_id/:topic_id", s.Handler.GetProducer)
	r.GET("v1/processor/:dispatcher_id", s.Handler.GetProcessor)

	s.server = eh_http.NewServer(&eh_http.ServerConfig{
		Addr: s.Opts.HTTPServerConfig.Addr,
	})
	s.server.SetHandler(r)
	if err := s.server.Serve(); err != nil {
		return err
	}
	return nil
}

// Stop stops the admin gracefully.
func (s *Admin) Stop() error {
	return s.server.Close()
}
