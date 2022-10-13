package prometheus

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentmusic/evhub/internal/processor/define"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/monitor/metric"

	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
)

// DispatchInterceptor is the dispatcher processing interceptor
func DispatchInterceptor(ctx context.Context, req interface{},
	info *handler.UnaryServerInfo, handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod == define.ProtocolTypeGRPCInternalHandler {
		return handler(ctx, req)
	}
	startTime := time.Now()
	// handle events
	iRsp, err := handler(ctx, req)
	// code to handle errors
	code := "0"
	if err != nil {
		// convert the error to code
		code = strconv.Itoa(int(errcode.ConvertRet(err).Code))
	}
	// convert request
	in, ok := req.(*define.SendReq)
	if !ok {
		log.Errorf("convert err")
		return iRsp, err
	}
	// calculating time spent
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	// metric report
	metric.MetricDispatcherClientRequestsTotal.With(prometheus.Labels{"code": code, "dispatcher_id": in.DispatchID}).Inc()
	metric.MetricDispatcherClientRequestDuration.With(prometheus.Labels{"dispatcher_id": in.DispatchID}).Observe(duration)
	metric.MetricDispatcherClientRetryTotal.With(prometheus.Labels{
		"dispatcher_id": in.DispatchID,
		"retry_times":   strconv.Itoa(int(in.RetryTimes)),
	}).Inc()
	return iRsp, err
}
