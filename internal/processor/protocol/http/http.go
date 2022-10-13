package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tencentmusic/evhub/internal/processor/define"
	"github.com/tencentmusic/evhub/internal/processor/options"
	eh_pc "github.com/tencentmusic/evhub/pkg/gen/proto/processor"
	"github.com/tencentmusic/evhub/pkg/util/addrutil"
)

// HTTP is a type of protocol
type HTTP struct {
	Opts *options.Options
}

// Send dispatch downstream service processing
func (s *HTTP) Send(ctx context.Context, req interface{}) (interface{}, error) {
	in, ok := req.(*define.SendReq)
	if !ok {
		return nil, fmt.Errorf("convert request error: %v", in)
	}
	// parse address
	addr := "http://" + addrutil.ParseAddr(in.EndPoint.Addr)
	httpClient := http.Client{
		Timeout: in.EndPoint.Timeout,
	}
	// make request
	dispatchReq := s.makeDispatchReq(in)
	b, err := json.Marshal(dispatchReq)
	if err != nil {
		return nil, err
	}
	// send request
	resp, err := httpClient.Post(addr, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var rsp eh_pc.DispatchRsp
	boydByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// unmarshal response
	if err := json.Unmarshal(boydByte, &rsp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code:%v err", resp.StatusCode)
	}
	return &define.SendRsp{
		PassBack: rsp.PassBack,
	}, nil
}

// makeDispatchReq assembly dispatch req
func (s *HTTP) makeDispatchReq(in *define.SendReq) *eh_pc.DispatchReq {
	return &eh_pc.DispatchReq{
		EventId:     in.EventID,
		Event:       in.Event,
		RetryTimes:  in.RetryTimes,
		RepeatTimes: in.RepeatTimes,
		PassBack:    in.PassBack,
	}
}
