package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tencentmusic/evhub/internal/producer/define"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	eh_pc "github.com/tencentmusic/evhub/pkg/gen/proto/processor"
	eh_http "github.com/tencentmusic/evhub/pkg/http"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Real-time event example
func main() {
	// report event
	if err := report("http://127.0.0.1:8083/v1/report", time.Second*5); err != nil {
		log.Errorf("report err:%v", err)
	}
	// start http server
	if err := startServer(":9008"); err != nil {
		log.Errorf("startServer err:%v", err)
	}
	select {}
}

// Start http server start the http service
func startServer(addr string) error {
	// initialize
	var r = gin.New()
	// handle
	r.POST("v1/dispatch", dispatch)
	server := eh_http.NewServer(&eh_http.ServerConfig{
		Addr: addr,
	})
	server.SetHandler(r)
	// start server
	if err := server.Serve(); err != nil {
		return err
	}
	return nil
}

// dispatch event distribution sink
func dispatch(c *gin.Context) {
	// req
	var req = &eh_pc.DispatchReq{}
	var rsp = &eh_pc.DispatchRsp{}
	defer c.JSON(http.StatusOK, &rsp)
	// parse req
	if err := c.ShouldBindJSON(req); err != nil {
		log.Errorf("should bind JSON err: %v", err)
		return
	}
	log.Infof("req:%+v", req)
}

// report event
func report(addr string, timeout time.Duration) error {
	// initialize http client
	httpClient := http.Client{
		Timeout: timeout,
	}
	req := &define.ReportReq{
		Event: &comm.Event{
			AppId:   "eh",
			TopicId: "test",
		},
		Trigger: &comm.EventTrigger{
			TriggerType: comm.EventTriggerType_EVENT_TRIGGER_TYPE_REAL_TIME,
		},
	}
	// marshal request
	reqByte, err := json.Marshal(req)
	if err != nil {
		return errors.Wrapf(err, "marshal")
	}
	// send request
	resp, err := httpClient.Post(addr, "application/octet-stream", bytes.NewBuffer(reqByte))
	if err != nil {
		return errors.Wrapf(err, "send post")
	}
	defer resp.Body.Close()
	// parse response
	var rsp define.ReportRsp
	boydByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "read all")
	}
	// unmarshal response
	if err := json.Unmarshal(boydByte, &rsp); err != nil {
		return errors.Wrapf(err, "unmarshal body:%v", string(boydByte))
	}
	// http status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code:%v err", resp.StatusCode)
	}
	log.Infof("rsp:%+v", rsp)
	return nil
}
