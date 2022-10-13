package log

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"go.uber.org/zap"
)

const (
	OutputStdout = iota
	OutputFile
	OutputStdoutAndFile
)

type EventKey string

const (
	EventAppID   EventKey = "event_app_id"
	EventTopicID EventKey = "event_topic_id"
	EventID      EventKey = "event_id"
	RetryTimes   EventKey = "retry_times"
	RepeatTimes  EventKey = "repeat_times"
)

const (
	IndexName = "index_name"
	AppIDName = "app_id"

	CallerSkipWith = -2
)

// Config is the configuration of the log.
type Config struct {
	Level    Level  // Level is the minimum enabled logging level.
	Output   int    // Output determines where the log should be written to.
	Filename string // Filename is the file to write logs to.
	MaxSize  int    // MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	MaxAge   int    // MaxAge is the maximum number of days to
	// retain old log files based on the timestamp encoded in their filename.
	MaxBackups   int  // MaxBackups is the maximum number of old log files to retain.
	Compress     bool // Compress determines if the rotated log files should be compressed using gzip.
	DevMode      bool // DevMode if true -> print colorful log in console and files.
	LogIndexName string
	AppID        string
}

var (
	defaultFilename   = "./log/default.log"
	defaultMaxSize    = 100
	defaultMaxAge     = 30
	defaultMaxBackups = 10
	defaultConfig     = &Config{Level: LevelInfo}

	LogIndexName = "evhub"
	AppID        = "evhub"
)

var (
	encoderConfigTimeKey = "time"
	extraFields          = []interface{}{IndexName, LogIndexName, AppIDName, AppID}
)

var logger = New(defaultConfig).With(extraFields...)

func GetLogger() Logger {
	return logger
}

// Init init the logger
func Init(config *Config) {
	if config.LogIndexName != "" {
		LogIndexName = config.LogIndexName
	}
	if config.AppID != "" {
		AppID = config.AppID
	}
	extraFields = []interface{}{IndexName, LogIndexName, AppIDName, AppID}
	logger = New(config).With(extraFields...)
}

// SetLevel alters the logging level.
func SetLevel(level Level) {
	logger.SetLevel(level)
}

// LevelHandler returns an HTTP handler that can dynamically modifies the log level.
func LevelHandler() http.Handler {
	return logger.LevelHandler()
}

// Debug logs a message with some additional context.
func Debug(msg string, keyvals ...interface{}) {
	logger.Debug(msg, keyvals...)
}

// Info logs a message with some additional context.
func Info(msg string, keyvals ...interface{}) {
	logger.Info(msg, keyvals...)
}

// Warn logs a message with some additional context.
func Warn(msg string, keyvals ...interface{}) {
	logger.Warn(msg, keyvals...)
}

// Error logs a message with some additional context.
func Error(msg string, keyvals ...interface{}) {
	logger.Error(msg, keyvals...)
}

// Panic logs a message with some additional context, then panics.
func Panic(msg string, keyvals ...interface{}) {
	logger.Panic(msg, keyvals...)
}

// Fatal logs a message with some additional context, then calls os.Exit.
func Fatal(msg string, keyvals ...interface{}) {
	logger.Fatal(msg, keyvals...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

// Close flushes any buffered log entries.
func Close() {
	logger.Close()
}

func With(ctx context.Context) *zap.SugaredLogger {
	withStr := []interface{}{string(EventAppID), ctx.Value(EventAppID), string(EventTopicID),
		ctx.Value(EventTopicID), string(EventID), ctx.Value(EventID), IndexName, LogIndexName, AppIDName, AppID,
		string(RetryTimes), ctx.Value(RetryTimes), string(RepeatTimes), ctx.Value(RepeatTimes)}
	return logger.WithOptions(zap.AddCallerSkip(CallerSkipWith)).With(withStr...)
}

func SetEventAppID(ctx context.Context, appID string) context.Context {
	return context.WithValue(ctx, EventAppID, appID)
}

func SetEventTopicID(ctx context.Context, topicID string) context.Context {
	return context.WithValue(ctx, EventTopicID, topicID)
}

func SetEventID(ctx context.Context, eventID string) context.Context {
	return context.WithValue(ctx, EventID, eventID)
}

func SetRetryTimes(ctx context.Context, retryTimes uint32) context.Context {
	return context.WithValue(ctx, RetryTimes, retryTimes)
}

func SetRepeatTimes(ctx context.Context, repeatTimes uint32) context.Context {
	return context.WithValue(ctx, RepeatTimes, repeatTimes)
}

func SetDye(ctx context.Context, appID string, topicID string,
	eventID string, retryTimes, repeatTimes uint32) context.Context {
	ctx = context.WithValue(ctx, EventAppID, appID)
	ctx = context.WithValue(ctx, EventTopicID, topicID)
	ctx = context.WithValue(ctx, RetryTimes, retryTimes)
	ctx = context.WithValue(ctx, RepeatTimes, repeatTimes)
	return context.WithValue(ctx, EventID, eventID)
}

func ShouldJSON(v interface{}) {
	withStr, err := makeArgs(v)
	if err != nil {
		logger.Infof("should json err:%v", err)
	}
	logger.WithOptions(zap.AddCallerSkip(-1)).With(withStr...).Info("shouldJSON")
}

func makeArgs(ptr interface{}) ([]interface{}, error) {
	if ptr == nil {
		return nil, fmt.Errorf("ptr is nil")
	}
	reType := reflect.TypeOf(ptr)
	if reType.Kind() != reflect.Ptr || reType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("ptr is not a pointer")
	}
	v := reflect.ValueOf(ptr).Elem()
	r := make([]interface{}, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		structField := v.Type().Field(i)
		tag := structField.Tag
		label := tag.Get("json")
		if label == "" {
			label = structField.Name
		}
		r = append(r, label, fmt.Sprintf("%+v", v.Field(i)))
	}
	return r, nil
}

func GetLevel() Level {
	return logger.GetLevel()
}
