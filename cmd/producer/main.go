package main

import (
	"github.com/tencentmusic/evhub/internal/producer"
	"github.com/tencentmusic/evhub/internal/producer/options"
	"github.com/tencentmusic/evhub/pkg/log"
	"github.com/tencentmusic/evhub/pkg/monitor"
	"github.com/tencentmusic/evhub/pkg/program"
	"github.com/tencentmusic/evhub/pkg/util/codeutil"
)

func main() {
	// create a configuration
	opts, err := options.NewOptions()
	if err != nil {
		log.Panicf("new options error: %v", err)
	}
	// init log
	log.Init(&opts.LogConfig)
	log.Infof("producer init with options: %s.\n", codeutil.GetUglyJsonStr(opts))
	// create new server of the producer
	s, err := producer.New(opts)
	if err != nil {
		log.Panicf("new producer error: %v", err)
	}
	// monitor
	monitor.Start(monitor.Address(opts.MonitorConfig.Addr))
	program.Run(s)
}
