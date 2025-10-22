package logger

import (
	"context"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

var globalLogger *Logger

func Init() error {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	globalLogger = &Logger{zapLogger}
	return nil
}

func InitDev() error {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	globalLogger = &Logger{zapLogger}
	return nil
}

func Get() *Logger {
	if globalLogger == nil {
		zapLogger, _ := zap.NewProduction()
		globalLogger = &Logger{zapLogger}
	}
	return globalLogger
}

func (l *Logger) WithCorrelationID(ctx context.Context, correlationID string) *Logger {
	return &Logger{l.With(zap.String("correlation_id", correlationID))}
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
