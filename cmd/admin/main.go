package main

import (
	"github.com/tencentmusic/evhub/internal/admin"
	"github.com/tencentmusic/evhub/internal/admin/options"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/monitor"
	"github.com/tencentmusic/evhub/pkg/program"
)

func main() {
	// create a configuration
	opts, err := options.NewOptions()
	if err != nil {
		log.Panicf("new options error: %v", err)
	}
	// init log
	log.Init(&opts.LogConfig)
	// create new server of the admin
	s, err := admin.New(opts)
	if err != nil {
		log.Panicf("new admin error: %v", err)
	}
	// monitor
	monitor.Start(monitor.Address(opts.MonitorConfig.Addr))
	// server started
	program.Run(s)
}
