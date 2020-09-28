package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	_logrus "github.com/sirupsen/logrus"
)

const ()

var (
	logger = &_logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(_logrus.TextFormatter),
		Hooks:        make(_logrus.LevelHooks),
		Level:        _logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	stdoutDupOnce, stderrDupOnce sync.Once
	isRuntimeCaller              bool
	timeFormat                   string
)

type Options struct {
	IsDebug bool

	// path and file name e.g. ../path/to/filename.log
	StdoutPath string
	StderrPath string

	EnableTimestamp bool
	TimestampFormat string // golang time format standard

	EnableRuntimeCaller bool
}

func Init(opts *Options) {
	if opts.StdoutPath != "" {
		stdoutDupOnce.Do(func() {
			set(opts.StdoutPath, int(os.Stdout.Fd()))
		})
	}
	logger.Formatter.(*_logrus.TextFormatter).DisableLevelTruncation = true
	logger.Formatter.(*_logrus.TextFormatter).DisableColors = false
	if opts.StderrPath != "" {
		stderrDupOnce.Do(func() {
			set(opts.StderrPath, int(os.Stderr.Fd()))
		})
		logger.Formatter.(*_logrus.TextFormatter).DisableColors = true
	}
	logger.Level = _logrus.InfoLevel
	if opts.IsDebug {
		logger.Level = _logrus.DebugLevel
	}
	if opts.EnableRuntimeCaller {
		logger.ReportCaller = true
		isRuntimeCaller = true
	}
	logger.Formatter.(*_logrus.TextFormatter).DisableTimestamp = true
	if opts.EnableTimestamp {
		logger.Formatter.(*_logrus.TextFormatter).DisableTimestamp = false
	}
	logger.Formatter.(*_logrus.TextFormatter).TimestampFormat = time.RFC3339
	if opts.TimestampFormat != "" {
		logger.Formatter.(*_logrus.TextFormatter).TimestampFormat = opts.TimestampFormat
		timeFormat = opts.TimestampFormat
	}
}

// set log to specified file with specified file descriptor.
func set(logPath string, fd int) {
	// check if log directory exist, if not create one.
	if _, err := os.Stat(filepath.Dir(logPath)); os.IsNotExist(err) {
		if err = os.Mkdir(filepath.Dir(logPath), 0744); err != nil {
			log.Panicf("log: error setting log path: %v", err)
		}
	}
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Panicf("log: failed opening log path: %v", err)
	}
	if err = syscall.Dup2(int(logFile.Fd()), fd); err != nil {
		log.Panicf("log: failed dup 2 to path %s with error: %v", logPath, err)
	}
}

func formatStdout(fileName, funcName string, line int) string {
	paths := strings.Split(fileName, "/")
	return fmt.Sprintf("%s:%s:%d", fmt.Sprintf("%s/%s", paths[len(paths)-2], paths[len(paths)-1]), funcName, line)
}

func Trace(msg ...interface{}) {
	logger.Trace(msg...)
}

func Debug(msg ...interface{}) {
	logger.Debug(msg...)
}

// Info prints to stdout
func Info(msg ...interface{}) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		source := formatStdout(file, runtime.FuncForPC(pc).Name(), line)
		if timeFormat == "" {
			timeFormat = time.RFC3339
		}
		if isRuntimeCaller {
			fmt.Printf("%s [INFO] %s src=%s\n", time.Now().Format(timeFormat), msg, source)
		} else {
			fmt.Printf("%s [INFO] %s\n", time.Now().Format(timeFormat), msg)
		}
	}
}

func Warn(msg ...interface{}) {
	logger.Warn(msg...)
}

func Error(err error, msg ...interface{}) {
	logger.WithFields(_logrus.Fields{"err": err}).Error(msg...)
}

func Fatal(err error, msg ...interface{}) {
	logger.WithFields(_logrus.Fields{"err": err}).Fatal(msg...)
}
