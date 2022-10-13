package log_test

import (
	"github.com/tencentmusic/evhub/pkg/log"
)

func Example() {
	log.Init(&log.Config{
		Level:      log.LevelInfo,
		Filename:   "test.log",
		MaxSize:    10,
		MaxAge:     1,
		MaxBackups: 2,
		Compress:   false,
		DevMode:    false,
	})

	people := "Alice"
	log.Debug("Hello", "people", people)
	log.Info("Hello", "people", people)
	log.Warn("Hello", "people", people)
	log.Error("Hello", "people", people)
}
