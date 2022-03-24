package log

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetupLogger_logger(t *testing.T) {
	logger = nil
	conf := LoggerConfig{
		Level: InfoLevel,
	}
	if err := SetupLogger(conf); err != nil {
		t.Errorf("error setting up logger: %s", err.Error())
	}
	if getLogger().Level != logrus.Level(InfoLevel) {
		t.Errorf("expected log level to be %s got %s", logrus.InfoLevel, getLogger().Level)
	}

	// TODO: how to test log format?
	/*
		loggerConfig.Format = TextFormat
		SetupLogger(loggerConfig)
		expected := &logrus.TextFormatter{}
		if getLogger().Formatter != expected {
			t.Errorf("expected formatter to be %+v got %+v", &logrus.TextFormatter{}, getLogger().Formatter)
		}
	*/

	conf = LoggerConfig{
		ReportCaller: true,
	}
	if err := SetupLogger(conf); err != nil {
		t.Errorf("error setting up logger: %s", err.Error())
	}
	if getLogger().ReportCaller == false {
		t.Errorf("expected ReportCaller to be %t got %t", true, getLogger().ReportCaller)
	}

	// TODO: test write output to file
}

func TestWriteMsgAllLevels_logger(t *testing.T) {
	logger = nil
	var buf bytes.Buffer
	getLogger().SetOutput(&buf)
	getLogger().SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	testMsg := "test"

	SetLogLevel(TraceLevel)
	expectedMsg := "level=trace msg=" + testMsg + "\n"
	Trace(testMsg)
	if buf.String() != expectedMsg {
		t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
	}
	buf.Reset()

	SetLogLevel(DebugLevel)
	expectedMsg = "level=debug msg=" + testMsg + "\n"
	Debug(testMsg)
	if buf.String() != expectedMsg {
		t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
	}
	buf.Reset()

	SetLogLevel(InfoLevel)
	expectedMsg = "level=info msg=" + testMsg + "\n"
	Info(testMsg)
	if buf.String() != expectedMsg {
		t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
	}
	buf.Reset()

	SetLogLevel(WarnLevel)
	expectedMsg = "level=warning msg=" + testMsg + "\n"
	Warn(testMsg)
	if buf.String() != expectedMsg {
		t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
	}
	buf.Reset()

	SetLogLevel(ErrorLevel)
	expectedMsg = "level=error msg=" + testMsg + "\n"
	Error(testMsg)
	if buf.String() != expectedMsg {
		t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
	}
	buf.Reset()

	SetLogLevel(PanicLevel)
	expectedMsg = "level=panic msg=" + testMsg + "\n"
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("error recovering from panic: %s", err)
		}
		if buf.String() != expectedMsg {
			t.Errorf("expected '%s' got '%s'", expectedMsg, buf.String())
		}
	}()
	Panic(testMsg)
	buf.Reset()

	// TODO: how to test fatal level log write (?)
}

// TestMultiRoutineAccess_logger tests that multiple goroutines are able to write log messages without creating multiple instances of logger.
func TestMultiRoutineAccess_logger(t *testing.T) {
	logger = nil
	var lock sync.Mutex
	loggerAddrs := make([]string, 0)
	var wgroup sync.WaitGroup

	for i := 0; i < 10; i++ {
		wgroup.Add(1)
		go func(index int) {
			defer wgroup.Done()

			log := getLogger()
			lock.Lock()
			loggerAddrs = append(loggerAddrs, fmt.Sprintf("%p", log))
			lock.Unlock()
			log.Info(fmt.Sprintf("This is a log entry from [%v]", index))
		}(i)
	}
	wgroup.Wait()

	if len(loggerAddrs) == 0 {
		t.Fatalf("expected at least one message")
	}
	origAddr := loggerAddrs[0]
	for i := 1; i < len(loggerAddrs); i++ {
		if origAddr != loggerAddrs[i] {
			t.Errorf("expected all logger references to share the same memory address")
		}
	}
}
