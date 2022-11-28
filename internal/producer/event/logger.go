package event

import (
	"context"
	"time"

	"github.com/tencentmusic/evhub/pkg/log"
	"gorm.io/gorm/logger"
)

type GormLogger struct{}

func (GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return GormLogger{}
}

func (GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	log.Infof(s, i...)
}

func (GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	log.Warnf(s, i...)
}

func (GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	log.Errorf(s, i...)
}

func (GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	//do nothing
}
