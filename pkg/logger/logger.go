package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

func New(logLevel string) *zap.Logger {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zap.InfoLevel
	}

	var config zap.Config
	switch level {
	case zap.DebugLevel:
		config = zap.NewDevelopmentConfig()
	case zap.InfoLevel:
		config = zap.NewProductionConfig()
	}

	_, err := os.Stat("logs")
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir("logs", os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile("logs/estimate.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	config.OutputPaths = append(config.OutputPaths, "logs/estimate.log")

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	return logger
}
