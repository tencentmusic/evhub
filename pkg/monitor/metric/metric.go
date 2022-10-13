package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// MetricRPCServerRequestsTotal is the total number of requests received by the server
	MetricRPCServerRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_server_requests_total",
		Help: "Total number of rpc requests made.",
	}, []string{"code", "method"})
	// MetricRPCServerRequestDuration is the duration of the request received by the server
	MetricRPCServerRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "rpc_server_request_duration_seconds",
		Help: "The rpc request latencies in seconds.",
	}, []string{"method"})
	// MetricRPCClientRequestsTotal is the total number of requests received by the client
	MetricRPCClientRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_client_requests_total",
		Help: "Total number of rpc requests made.",
	}, []string{"code", "method"})
	// MetricRPCClientRequestDuration 	 is the total number of requests received by the client
	MetricRPCClientRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "rpc_client_request_duration_seconds",
		Help: "The rpc request latencies in seconds.",
	}, []string{"method"})
	// MetricDispatcherClientRequestsTotal is the total number of requests received by the dispatcher
	MetricDispatcherClientRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatcher_client_requests_total",
		Help: "Total number of dispatcher requests made.",
	}, []string{"code", "dispatcher_id"})
	// MetricDispatcherClientRequestDuration is the total number of requests received by the dispatcher
	MetricDispatcherClientRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dispatcher_client_request_duration_seconds",
		Help: "The dispatcher request latencies in seconds.",
	}, []string{"dispatcher_id"})
	// MetricDispatcherClientRetryTotal is the total number of requests received by the dispatcher
	MetricDispatcherClientRetryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatcher_client_retry_total",
		Help: "Total number of dispatcher retry made.",
	}, []string{"dispatcher_id", "retry_times"})
	// MetricProducerClientRequestsTotal is the total number of requests received by the producer
	MetricProducerClientRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "producer_client_requests_total",
		Help: "Total number of producer made.",
	}, []string{"app_id", "topic_id", "interface", "code"})
	// MetricProducerClientRequestDuration is the total number of requests received by the producer
	MetricProducerClientRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "producer_client_request_duration_seconds",
		Help: "The producer request latencies in seconds.",
	}, []string{"app_id", "topic_id", "interface"})
	// MetricCallbackClientRetryTotal is the total number of requests received by the client
	MetricCallbackClientRetryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "callback_client_retry_total",
		Help: "Total number of callback retry made.",
	}, []string{"app_id", "topic_id", "retry_times"})
)

func init() {
	prometheus.MustRegister(MetricRPCServerRequestsTotal)
	prometheus.MustRegister(MetricRPCServerRequestDuration)
	prometheus.MustRegister(MetricRPCClientRequestsTotal)
	prometheus.MustRegister(MetricRPCClientRequestDuration)
	prometheus.MustRegister(MetricDispatcherClientRequestsTotal)
	prometheus.MustRegister(MetricDispatcherClientRequestDuration)
	prometheus.MustRegister(MetricDispatcherClientRetryTotal)
	prometheus.MustRegister(MetricProducerClientRequestsTotal)
	prometheus.MustRegister(MetricProducerClientRequestDuration)
	prometheus.MustRegister(MetricCallbackClientRetryTotal)
}
