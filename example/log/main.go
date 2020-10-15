package main

import (
	"fmt"
	"log"
	"time"

	_log "github.com/mfathirirhas/godevkit/log"
)

func main() {
	_log.Init(&_log.Options{
		StdoutPath:          `/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/godevkit/example/log/test_log/stdout/out.log`,
		StderrPath:          `/Users/mfathirirhas/code/go/src/github.com/mfathirirhas/godevkit/example/log/test_log/stderr/error.log`,
		EnableRuntimeCaller: true,
		EnableTimestamp:     true,
		IsDebug:             true,
		TimestampFormat:     time.RFC822,
	})
	_log.Trace("Trace message")
	_log.Debug("Debug message")
	_log.Info("Info message")
	_log.Warn("Warn message")
	fmt.Println("print to stdout")
	fmt.Println("print to stdout2")
	log.Println("print to stderr")
}
