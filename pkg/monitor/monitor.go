package monitor

import (
	"net/http"
	_ "net/http/pprof" // pprof

	"github.com/tencentmusic/evhub/pkg/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Options struct {
	Addr string
}

type Option func(o *Options)

// Address sets the address of the monitor.
func Address(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// Start starts the prometheus exporter.
func Start(opts ...Option) {
	options := &Options{
		Addr: ":8080",
	}
	for _, o := range opts {
		o(options)
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(options.Addr, nil); err != nil {
			log.Error("failed to start monitor", "error", err)
		}
	}()
}
