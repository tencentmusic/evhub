package http

import (
	"context"
	"net/http"

	"github.com/tencentmusic/evhub/internal/producer/define"

	"github.com/gin-gonic/gin"
	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/handler"
	eh_http "github.com/tencentmusic/evhub/pkg/http"
)

// initialize
func init() {
	gin.SetMode(gin.ReleaseMode)
}

// HTTP is a type of protocol
type HTTP struct {
	H      *handler.Handler
	Opts   *options.Options
	server *eh_http.Server
}

// Start initialize some operations
func (s *HTTP) Start() error {
	var r = gin.New()
	r.POST("v1/report", s.report)
	r.POST("v1/prepare", s.prepare)
	r.POST("v1/commit", s.commit)
	r.POST("v1/rollback", s.rollback)
	s.server = eh_http.NewServer(&eh_http.ServerConfig{
		Addr: s.Opts.HTTPServerConfig.Addr,
	})
	s.server.SetHandler(r)
	if err := s.server.Serve(); err != nil {
		return err
	}
	return nil
}

// makeReq is used to make a request to the server
func (s *HTTP) makeReq(c *gin.Context, req interface{}) *comm.Result {
	if err := c.ShouldBindJSON(req); err != nil {
		return &comm.Result{
			Code: errcode.CommParamInvalid,
			Msg:  err.Error(),
		}
	}
	return nil
}

// handle is used to handle a request
func (s *HTTP) handle(ctx context.Context, req interface{}) (interface{}, *comm.Result) {
	rsp, err := s.H.Handler(ctx, req,
		&handler.UnaryServerInfo{FullMethod: define.ReportEventReport})
	if err != nil {
		return rsp, &comm.Result{
			Code: errcode.SystemErr,
			Msg:  err.Error(),
		}
	}
	return rsp, nil
}

// report is used to report events to the server
func (s *HTTP) report(c *gin.Context) {
	var req = &define.ReportReq{}
	var rsp = &define.ReportRsp{}
	defer c.JSON(http.StatusOK, &rsp)

	//parse request
	if ret := s.makeReq(c, req); ret != nil {
		rsp.Ret = ret
		return
	}

	//handle
	iRsp, ret := s.handle(c, req)
	if ret != nil {
		rsp.Ret = ret
		return
	}

	//response
	switch iRsp.(type) {
	case *define.ReportRsp:
		rsp = iRsp.(*define.ReportRsp)
	default:
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  "convert error",
		}
	}
}

// prepare is used to prepare a transaction message to the server
func (s *HTTP) prepare(c *gin.Context) {
	var req = &define.PrepareReq{}
	var rsp = &define.PrepareRsp{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	//parse request
	if ret := s.makeReq(c, req); ret != nil {
		rsp.Ret = ret
		return
	}

	//handle
	iRsp, ret := s.handle(c, req)
	if ret != nil {
		rsp.Ret = ret
		return
	}

	//response
	switch iRsp.(type) {
	case *define.PrepareRsp:
		rsp = iRsp.(*define.PrepareRsp)
	default:
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  "convert error",
		}
	}
}

// commit is used to commit a transaction message to the server
func (s *HTTP) commit(c *gin.Context) {
	var req = &define.CommitReq{}
	var rsp = &define.CommitRsp{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	//parse request
	if ret := s.makeReq(c, req); ret != nil {
		rsp.Ret = ret
		return
	}

	//handle
	iRsp, ret := s.handle(c, req)
	if ret != nil {
		rsp.Ret = ret
		return
	}

	//response
	switch iRsp.(type) {
	case *define.CommitRsp:
		rsp = iRsp.(*define.CommitRsp)
	default:
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  "convert error",
		}
	}
}

// rollback is used to roll back a transaction message to the server
func (s *HTTP) rollback(c *gin.Context) {
	var req = &define.RollbackReq{}
	var rsp = &define.RollbackRsp{}
	defer func() {
		c.JSON(http.StatusOK, &rsp)
	}()

	//parse request
	if ret := s.makeReq(c, req); ret != nil {
		rsp.Ret = ret
		return
	}

	//handle
	iRsp, ret := s.handle(c, req)
	if ret != nil {
		rsp.Ret = ret
		return
	}

	//response
	switch iRsp.(type) {
	case *define.RollbackRsp:
		rsp = iRsp.(*define.RollbackRsp)
	default:
		rsp.Ret = &comm.Result{
			Code: errcode.SystemErr,
			Msg:  "convert error",
		}
	}
}

//Stop stops the http server gracefully.
func (s *HTTP) Stop() error {
	return s.server.Close()
}
