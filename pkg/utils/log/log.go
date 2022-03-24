package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type LoggerConfig struct {
	Level         LogLevel
	Format        LogFormat
	ReportCaller  bool
	OutputFiles   []string
	WriteToStdOut bool
}

type LogFormat int

const (
	TextFormat = iota
	JSONFormat
)

func SetupLogger(conf LoggerConfig) error {
	getLogger().Level = logrus.Level(conf.Level)

	switch conf.Format {
	case TextFormat:
		getLogger().SetFormatter(&logrus.TextFormatter{})
	case JSONFormat:
		getLogger().SetFormatter(&logrus.JSONFormatter{})
	default:
		return fmt.Errorf("invalid log format")
	}

	if conf.ReportCaller {
		getLogger().SetReportCaller(true) // NOTE: This will cause noticeable overhead
	}

	var writers []io.Writer
	if conf.WriteToStdOut {
		writers = append(writers, os.Stdout)
	}
	if len(conf.OutputFiles) != 0 {
		for _, path := range conf.OutputFiles {
			file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return fmt.Errorf("could not open log file '%s': %s", path, err.Error())
			}
			writers = append(writers, file)
		}
	}
	if len(writers) == 0 {
		return errors.New("there must be at least one output for logging")
	}
	return nil
}

type LogLevel int

const (
	TraceLevel LogLevel = LogLevel(logrus.TraceLevel)
	DebugLevel          = LogLevel(logrus.DebugLevel)
	InfoLevel           = LogLevel(logrus.InfoLevel)
	WarnLevel           = LogLevel(logrus.WarnLevel)
	ErrorLevel          = LogLevel(logrus.ErrorLevel)
	PanicLevel          = LogLevel(logrus.PanicLevel)
	FatalLevel          = LogLevel(logrus.FatalLevel)
)

func SetLogLevel(level LogLevel) {
	getLogger().SetLevel(logrus.Level(level))
}

func Trace(args ...interface{}) {
	getLogger().Trace(args...)
}

func Debug(args ...interface{}) {
	getLogger().Debug(args...)
}

func Info(args ...interface{}) {
	getLogger().Info(args...)
}

func Warn(args ...interface{}) {
	getLogger().Warn(args...)
}

func Error(args ...interface{}) {
	getLogger().Error(args...)
}

func Panic(args ...interface{}) {
	getLogger().Panic(args...)
}

func Fatal(args ...interface{}) {
	getLogger().Fatal(args...)
}

var lock sync.Mutex
var logger *logrus.Logger

func getLogger() *logrus.Logger {
	lock.Lock()
	defer lock.Unlock()

	if logger == nil {
		logger = logrus.New()
	}
	return logger
}
