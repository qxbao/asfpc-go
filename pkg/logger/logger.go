package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitLogger(development bool) error {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		return err
	}
	Logger = logger.Sugar()
	return nil
}

func GetLogger(loggerName *string) *zap.SugaredLogger {
	if Logger == nil {
		logger, _ := zap.NewDevelopment()
		Logger = logger.Sugar()
	}
	if loggerName == nil {
		return Logger
	}
	return Logger.Named(*loggerName)
}

func FlushLogger() {
	if Logger != nil {
		Logger.Sync()
	}
}