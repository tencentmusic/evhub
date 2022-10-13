package log

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init(&Config{Level: LevelFatal})
	Debug("test")
	Debugf("%s", "test")
	Info("test")
	Infof("%s", "test")
	Warn("test")
	Warnf("%s", "test")
	Error("test")
	Errorf("%s", "test")
	assert.Panics(t, func() { Panic("test") })
	assert.Panics(t, func() { Panicf("%s", "test") })
	Close()
}

func TestNew(t *testing.T) {
	logger := New(defaultConfig)
	logger.SetLevel(LevelFatal)
	logger.Debug("test")
	logger.Debugf("%s", "test")
	logger.Info("test")
	logger.Infof("%s", "test")
	logger.Warn("test")
	logger.Warnf("%s", "test")
	logger.Error("test")
	logger.Errorf("%s", "test")
	assert.Panics(t, func() { logger.Panic("test") })
	assert.Panics(t, func() { logger.Panicf("%s", "test") })
	logger.Close()
}

func TestNewOutput(t *testing.T) {
	logger := New(&Config{})
	logger.Info("test")

	logger = New(&Config{Output: OutputStdout})
	logger.Info("test")

	logger = New(&Config{Output: OutputFile})
	logger.Info("test")

	logger = New(&Config{Output: OutputStdoutAndFile})
	logger.Info("test")
}

func TestNewNil(t *testing.T) {
	New(nil).Debug("test")
}

func TestDevMode(t *testing.T) {
	New(&Config{Level: LevelInfo, DevMode: true}).Debug("test")
}

func TestSetLevel(t *testing.T) {
	SetLevel(LevelInfo)
	Debug("test")
}

func TestLevelHandler(t *testing.T) {
	assert.Implements(t, (*http.Handler)(nil), LevelHandler())
}

func BenchmarkLog(b *testing.B) {
	logger := New(defaultConfig)
	logger.SetLevel(LevelError)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test")
	}
}

func BenchmarkLogWithField(b *testing.B) {
	logger := New(defaultConfig)
	logger.SetLevel(LevelError)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test", "foo", "bar")
	}
}
