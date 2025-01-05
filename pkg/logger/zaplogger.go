package logger

import (
	"log"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var m sync.Mutex
var globalLogger atomic.Value // use Logger for type-safe and performance

func init() {
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         "json",
		OutputPaths:      []string{"stdout", "server.log"},
		ErrorOutputPaths: []string{"stderr", "server.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "msg",
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	m.Lock()
	defer m.Unlock()
	globalLogger.Store(logger)
}

func ll() *zap.Logger {
	return globalLogger.Load().(*zap.Logger)
}

func Info(msg string, fields ...zap.Field) {
	ll().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	ll().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	ll().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	ll().Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	ll().Fatal(msg, fields...)
}

func Sync() error {
	return ll().Sync()
}
