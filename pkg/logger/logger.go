package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface defines logging methods
type Logger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) *logrus.Entry
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
	
	return log
}