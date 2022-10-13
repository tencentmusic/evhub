package log

import (
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the fundamental interface for all log operations.
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	Panic(msg string, keyvals ...interface{})
	Fatal(msg string, keyvals ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	With(args ...interface{}) Logger

	WithOptions(opts ...zap.Option) *zap.SugaredLogger

	SetLevel(Level)
	LevelHandler() http.Handler

	Close() error
	GetLevel() Level
}

// New returns a Logger instance.
func New(config *Config) Logger {
	if config == nil {
		config = defaultConfig
	}

	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = encoderConfigTimeKey
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if config.DevMode {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	if config.Filename == "" {
		config.Filename = defaultFilename
	}
	if config.MaxSize <= 0 {
		config.MaxSize = defaultMaxSize
	}
	if config.MaxAge <= 0 {
		config.MaxAge = defaultMaxAge
	}
	if config.MaxBackups <= 0 {
		config.MaxBackups = defaultMaxBackups
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		LocalTime:  true,
		Compress:   config.Compress,
	}
	var writeSyncer zapcore.WriteSyncer
	switch config.Output {
	case OutputFile:
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger))
	case OutputStdoutAndFile:
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	default:
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}

	level := zap.NewAtomicLevelAt(config.Level.ZapLevel())
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller(), zap.AddCallerSkip(2))

	return &zapLogger{config: config, logger: logger.Sugar(), level: level, zapLogger: logger}
}

type zapLogger struct {
	config    *Config
	logger    *zap.SugaredLogger
	level     zap.AtomicLevel
	zapLogger *zap.Logger
}

func (l *zapLogger) WithOptions(opts ...zap.Option) *zap.SugaredLogger {
	l.checkLevel()
	return l.zapLogger.WithOptions(opts...).Sugar()
}

func (l *zapLogger) With(args ...interface{}) Logger {
	l.checkLevel()
	l.logger = l.logger.With(args...)
	return l
}

func (l *zapLogger) SetLevel(lvl Level) {
	l.level.SetLevel(lvl.ZapLevel())
	l.config.Level = lvl
}

func (l *zapLogger) checkLevel() {
	if l.config.Level.ZapLevel() != l.level.Level() {
		l.SetLevel(l.config.Level)
	}
}

func (l *zapLogger) LevelHandler() http.Handler {
	return l.level
}

// Debug logs a message with some additional context.
func (l *zapLogger) Debug(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Debugw(msg, keyvals...)
}

// Info logs a message with some additional context.
func (l *zapLogger) Info(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Infow(msg, keyvals...)
}

// Warn logs a message with some additional context.
func (l *zapLogger) Warn(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Warnw(msg, keyvals...)
}

// Error logs a message with some additional context.
func (l *zapLogger) Error(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Errorw(msg, keyvals...)
}

// Panic logs a message with some additional context, then panics.
func (l *zapLogger) Panic(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Panicw(msg, keyvals...)
}

// Fatal logs a message with some additional context, then calls os.Exit.
func (l *zapLogger) Fatal(msg string, keyvals ...interface{}) {
	l.checkLevel()
	l.logger.Fatalw(msg, keyvals...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (l *zapLogger) Panicf(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.checkLevel()
	l.logger.Fatalf(template, args...)
}

// Close flushes any buffered log entries.
func (l *zapLogger) Close() error {
	return l.logger.Sync()
}

// GetLevel is used to get the log level
func (l *zapLogger) GetLevel() Level {
	return l.config.Level
}
