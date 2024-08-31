package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func InitLogger() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger")
	}
	Log = logger.Sugar()
}

func init() {
	InitLogger()
}
