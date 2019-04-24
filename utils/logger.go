package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	switch os.Getenv("LOG_LEVEL") {
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}

	log.SetFormatter(formatter)
}

func Debug(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func Error(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
