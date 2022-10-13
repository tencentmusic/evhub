package interceptor

import (
	"github.com/tencentmusic/evhub/internal/producer/interceptor/prometheus"
	"github.com/tencentmusic/evhub/pkg/handler"
)

// Register is used to register
func Register() []handler.Interceptor {
	// report interceptor、prepare interceptor、commit interceptor、
	// roll back interceptor、internal handler interceptor、callback interceptor
	return []handler.Interceptor{prometheus.ReportInterceptor, prometheus.PrepareInterceptor,
		prometheus.CommitInterceptor, prometheus.RollbackInterceptor,
		prometheus.InternalHandlerInterceptor, prometheus.CallBackInterceptor, prometheus.LogDebugInterceptor}
}
