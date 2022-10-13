package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestMatchLogLevel(t *testing.T) {
	tests := []struct {
		level Level
		want  zapcore.Level
	}{
		{LevelDebug, zapcore.DebugLevel},
		{LevelInfo, zapcore.InfoLevel},
		{LevelWarn, zapcore.WarnLevel},
		{LevelError, zapcore.ErrorLevel},
		{LevelPanic, zapcore.PanicLevel},
		{LevelFatal, zapcore.FatalLevel},
		{Level(-1), zapcore.InfoLevel},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.level.ZapLevel())
	}
}
