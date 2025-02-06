package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(cfgLogLevel, cfgLogOutput, cfgLogFilePath string) *logrus.Logger {
	logger := logrus.New()

	switch cfgLogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	switch cfgLogOutput {
	case "file":
		file, err := os.OpenFile(cfgLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalf("Ошибка открытия файла логов: %v", err)
		}
		logger.SetOutput(file)
	default:
		logger.SetOutput(os.Stdout)
	}

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return logger
}
