package observability

import (
	"context"
	goaiObs "github.com/shaharia-lab/goai/observability"
	"github.com/sirupsen/logrus"
)

// LogrusLogger implements the Logger interface using logrus
type LogrusLogger struct {
	logger *logrus.Entry
}

// NewLogrusLogger creates a new LogrusLogger instance
func NewLogrusLogger(log *logrus.Logger) goaiObs.Logger {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return &LogrusLogger{
		logger: logrus.NewEntry(log),
	}
}

// Formatting methods
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *LogrusLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panicf(format, args...)
}

// Regular logging methods
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *LogrusLogger) Panic(args ...interface{}) {
	l.logger.Panic(args...)
}

// WithFields adds structured fields to the logger
func (l *LogrusLogger) WithFields(fields map[string]interface{}) goaiObs.Logger {
	return &LogrusLogger{
		logger: l.logger.WithFields(logrus.Fields(fields)),
	}
}

// WithContext adds context to the logger
func (l *LogrusLogger) WithContext(ctx context.Context) goaiObs.Logger {
	return &LogrusLogger{
		logger: l.logger.WithContext(ctx),
	}
}

// WithErr adds an error to the logger
func (l *LogrusLogger) WithErr(err error) goaiObs.Logger {
	return &LogrusLogger{
		logger: l.logger.WithError(err),
	}
}
