package interceptor

import (
	"github.com/tencentmusic/evhub/internal/processor/interceptor/prometheus"
	"github.com/tencentmusic/evhub/pkg/handler"
)

// Register registering interceptors
func Register() []handler.Interceptor {
	return []handler.Interceptor{prometheus.DispatchInterceptor}
}
