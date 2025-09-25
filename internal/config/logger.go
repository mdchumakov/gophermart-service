package config

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerInterface определяет общий интерфейс для логгеров
type LoggerInterface interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Sync() error
}

// NewLogger создает логгер в зависимости от окружения
func NewLogger(isProd bool) (LoggerInterface, error) {
	if isProd {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}
		return logger.Sugar(), nil
	}

	return NewDevLogger()
}

// NewDevLogger создает логгер для разработки
func NewDevLogger() (LoggerInterface, error) {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.TimeKey = "T"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger.Sugar(), nil
}

// SyncLogger синхронизирует логгер
func SyncLogger(logger LoggerInterface) {
	if logger != nil {
		if err := logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "bad file descriptor") {
				logger.Error("Failed to sync logger", "error", err)
			}
		}
	}
}
