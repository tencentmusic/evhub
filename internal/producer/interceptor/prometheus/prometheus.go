package prometheus

import (
	"context"
	"strconv"
	"time"

	"github.com/tencentmusic/evhub/internal/producer/define"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentmusic/evhub/pkg/errcode"
	"github.com/tencentmusic/evhub/pkg/gen/proto/comm"
	"github.com/tencentmusic/evhub/pkg/monitor/metric"
	"google.golang.org/grpc/metadata"

	"github.com/tencentmusic/evhub/pkg/handler"
	"github.com/tencentmusic/evhub/pkg/log"
)

// LogDebugInterceptor
func LogDebugInterceptor(ctx context.Context, req interface{},
	info *handler.UnaryServerInfo, handler handler.UnaryHandler) (interface{}, error) {
	iRsp, err := handler(ctx, req)
	log.Debugf("access req:%+v rsp:%+v", req, iRsp)
	return iRsp, err
}

// ReportInterceptor is used to report metrics for the report function
func ReportInterceptor(ctx context.Context, req interface{},
	info *handler.UnaryServerInfo, handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.ReportEventReport {
		return handler(ctx, req)
	}
	// report event
	startTime := time.Now()
	// handle events
	iRsp, err := handler(ctx, req)
	code := "0"
	// convert response
	rsp, ok := iRsp.(*define.ReportRsp)
	if err != nil || !ok {
		// system error
		code = strconv.Itoa(errcode.SystemErr)
	}
	if ok && rsp.Ret != nil {
		// handle error
		code = strconv.Itoa(int(rsp.Ret.Code))
	}
	// calculating time spent
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	// convert request
	in, ok := req.(*define.ReportReq)
	if !ok {
		log.Errorf("convert err")
		return iRsp, err
	}
	// report metrics
	makeProducerClientMetric(in.Event.AppId,
		in.Event.TopicId, info.FullMethod, code, duration)
	return iRsp, err
}

// PrepareInterceptor is used to report metrics for the prepare function
func PrepareInterceptor(ctx context.Context, req interface{},
	info *handler.UnaryServerInfo, handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.ReportEventPrepare {
		return handler(ctx, req)
	}
	//transaction prepare
	startTime := time.Now()
	// handle transaction prepared events
	iRsp, err := handler(ctx, req)
	code := "0"
	// convert request
	rsp, ok := iRsp.(*define.PrepareRsp)
	if err != nil || !ok {
		code = strconv.Itoa(errcode.SystemErr)
	}
	if ok && rsp.Ret != nil {
		// handler err
		code = strconv.Itoa(int(rsp.Ret.Code))
	}
	// calculating time spent
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	// convert request
	in, ok := req.(*define.PrepareReq)
	if !ok {
		log.Errorf("convert err")
		return iRsp, err
	}
	// report metrics
	makeProducerClientMetric(in.Event.AppId, in.Event.TopicId,
		info.FullMethod, code, duration)

	return iRsp, err
}

// CommitInterceptor is used to report metrics for the commit function
func CommitInterceptor(ctx context.Context, req interface{}, info *handler.UnaryServerInfo,
	handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.ReportEventCommit {
		return handler(ctx, req)
	}
	startTime := time.Now()
	//transaction commit
	iRsp, err := handler(ctx, req)
	code := "0"
	// convert response
	rsp, ok := iRsp.(*define.CommitRsp)
	if err != nil || !ok {
		code = strconv.Itoa(errcode.SystemErr)
	}
	if ok && rsp.Ret != nil {
		// handler err
		code = strconv.Itoa(int(rsp.Ret.Code))
	}
	// calculating time spent
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	appID, topicID := fromContext(ctx)
	// report metrics
	makeProducerClientMetric(appID, topicID, info.FullMethod,
		code, duration)

	return iRsp, err
}

// RollbackInterceptor is the used report metric for the rollback function
func RollbackInterceptor(ctx context.Context, req interface{}, info *handler.UnaryServerInfo,
	handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.ReportEventRollback {
		return handler(ctx, req)
	}
	startTime := time.Now()
	//transaction rollback
	iRsp, err := handler(ctx, req)
	code := "0"
	// convert response
	rsp, ok := iRsp.(*define.RollbackRsp)
	if err != nil || !ok {
		code = strconv.Itoa(errcode.SystemErr)
	}
	if ok && rsp.Ret != nil {
		// handler err
		code = strconv.Itoa(int(rsp.Ret.Code))
	}
	// calculating time spent
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	appID, topicID := fromContext(ctx)
	// report metrics
	makeProducerClientMetric(appID, topicID, info.FullMethod,
		code, duration)
	return iRsp, err
}

// InternalHandlerInterceptor  is used to report metrics for the internal handler function
func InternalHandlerInterceptor(ctx context.Context, req interface{}, info *handler.UnaryServerInfo,
	handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.InternalHandler {
		return handler(ctx, req)
	}
	iRsp, code, duration, err := makeCodeDuration(ctx, req, handler)

	in, ok := req.(*define.InternalHandlerReq)
	if !ok {
		log.Errorf("convert err")
		return iRsp, err
	}
	appID, topicID := in.EvhubMsg.Event.AppId, in.EvhubMsg.Event.TopicId

	makeProducerClientMetric(appID, topicID, info.FullMethod, code, duration)
	return iRsp, err
}

// CallBackInterceptor  is used to report metrics for the callback function
func CallBackInterceptor(ctx context.Context, req interface{}, info *handler.UnaryServerInfo,
	handler handler.UnaryHandler) (interface{}, error) {
	if info.FullMethod != define.InternalHandler {
		return handler(ctx, req)
	}
	iRsp, code, duration, err := makeCodeDuration(ctx, req, handler)

	in, ok := req.(*define.InternalHandlerReq)
	if !ok {
		log.Errorf("convert err")
		return iRsp, err
	}
	if in.EvhubMsg.Status != comm.EventStatus_EVENT_STATUS_TX_CALLBACK_IN_DELAY_QUEUE {
		return iRsp, err
	}

	appID, topicID := in.EvhubMsg.Event.AppId, in.EvhubMsg.Event.TopicId

	makeProducerClientMetric(appID, topicID, define.ReportEventCallback, code, duration)

	metric.MetricCallbackClientRetryTotal.With(prometheus.Labels{
		"app_id":      appID,
		"topic_id":    topicID,
		"retry_times": strconv.Itoa(int(in.EvhubMsg.DispatcherInfo.RetryTimes)),
	}).Inc()
	return iRsp, err
}

// fromContext returns appID, topicID
func fromContext(ctx context.Context) (string, string) {
	appID, topicID := "", ""
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ""
	}
	t := md.Get(define.EventAppID)
	if len(t) > 0 {
		appID = t[0]
	}
	t = md.Get(define.EventTopicID)
	if len(t) > 0 {
		topicID = t[0]
	}

	return appID, topicID
}

//makeCodeDuration returns code、duration
func makeCodeDuration(ctx context.Context, req interface{}, handler handler.UnaryHandler) (interface{},
	string, float64, error) {
	startTime := time.Now()
	iRsp, err := handler(ctx, req)
	code := "0"
	rsp, ok := iRsp.(*define.InternalHandlerRsp)
	if err != nil || !ok {
		code = strconv.Itoa(errcode.SystemErr)
	}
	if ok && rsp.Ret != nil {
		code = strconv.Itoa(int(rsp.Ret.Code))
	}
	duration := float64(time.Since(startTime).Milliseconds()) / 1000
	return iRsp, code, duration, err
}

// makeProducerClientMetric is used to report metrics
func makeProducerClientMetric(appID string,
	topicID string, method string, code string, duration float64) {
	metric.MetricProducerClientRequestsTotal.With(prometheus.Labels{"app_id": appID,
		"topic_id": topicID, "interface": method, "code": code}).Inc()
	metric.MetricProducerClientRequestDuration.With(prometheus.Labels{"app_id": appID,
		"topic_id": topicID, "interface": method}).Observe(duration)
}
