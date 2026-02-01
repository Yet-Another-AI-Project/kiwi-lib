package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/otelutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ContextKey string

const HostIDKey ContextKey = "host_id"
const TaskIDKey ContextKey = "task_id"

type Logger struct {
	level         string
	format        string
	filePath      string
	DefaultLogger *zap.Logger
}

func (log *Logger) getFieldsFromContext(ctx context.Context) []zap.Field {
	fields := []zap.Field{}
	traceID := otelutils.GetTraceID(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	return fields
}

func (log *Logger) Debugf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Debug(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) Infof(ctx context.Context, msg string, args ...interface{}) {

	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Info(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) Warnf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Warn(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) Errorf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Error(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) DPanicf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.DPanic(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) Panicf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Panic(fmt.Sprintf(msg, args...), fields...)
}

func (log *Logger) Fatalf(ctx context.Context, msg string, args ...interface{}) {
	fields := log.getFieldsFromContext(ctx)

	log.DefaultLogger.Fatal(fmt.Sprintf(msg, args...), fields...)
}

func newDefaultLogger(level string, format string, filePath string) (*zap.Logger, error) {

	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(pe)

	switch format {
	case "json":
		encoder = zapcore.NewJSONEncoder(pe)
	case "console":
		encoder = zapcore.NewConsoleEncoder(pe)
	}

	var syncer zapcore.WriteSyncer
	if filePath != "" {
		syncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    500,
			MaxBackups: 100,
			MaxAge:     7,
			Compress:   true,
		})
	} else {
		syncer = zapcore.AddSync(os.Stdout)
	}

	zaplevel := zap.InfoLevel

	switch level {
	case "debug":
		zaplevel = zap.DebugLevel
	case "info":
		zaplevel = zap.InfoLevel
	case "error":
		zaplevel = zap.ErrorLevel
	case "dev":
		zaplevel = zap.DPanicLevel
	}

	core := zapcore.NewCore(encoder, syncer, zap.NewAtomicLevelAt(zaplevel))

	logger := zap.New(core, zap.WithCaller(false))

	zap.ReplaceGlobals(logger)

	return logger, nil
}

type Option func(*Logger)

func WithLevel(level string) Option {
	return func(logger *Logger) {
		logger.level = level
	}
}

func WithFilePath(filePath string) Option {
	return func(logger *Logger) {
		logger.filePath = filePath
	}
}

func WithFormat(format string) Option {
	return func(logger *Logger) {
		logger.format = format
	}
}

func NewLogger(opts ...Option) (*Logger, error) {
	logger := &Logger{}

	for _, opt := range opts {
		opt(logger)
	}

	defaultLogger, err := newDefaultLogger(logger.level, logger.format, logger.filePath)
	if err != nil {
		return nil, err
	}

	logger.DefaultLogger = defaultLogger
	return logger, nil
}
