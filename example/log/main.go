package main

import (
	"time"

	_log "github.com/mfathirirhas/godevkit/log"
)

func main() {
	_log.Init(&_log.Options{
		StdoutPath:          "./example/log/test_log/stdout/out.log",
		StderrPath:          "./example/log/test_log/stderr/error.log",
		EnableRuntimeCaller: true,
		EnableTimestamp:     true,
		IsDebug:             true,
		TimestampFormat:     time.RFC822,
	})
	_log.Trace("Trace message")
	_log.Debug("Debug message")
	_log.Info("Info message")
	_log.Warn("Warn message")
}
