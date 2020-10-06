package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_logrus "github.com/sirupsen/logrus"
)

var (
	logger = &_logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(_logrus.TextFormatter),
		Hooks:        make(_logrus.LevelHooks),
		Level:        _logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	dupOnce         sync.Once
	isRuntimeCaller bool
	timeFormat      string
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
	dupOnce.Do(func() {
		if opts.StdoutPath != "" {
			set(opts.StdoutPath, int(os.Stdout.Fd()))
		}
		if opts.StderrPath != "" {
			set(opts.StderrPath, int(os.Stderr.Fd()))
		}
	})

	logger.Formatter.(*_logrus.TextFormatter).DisableLevelTruncation = true
	logger.Formatter.(*_logrus.TextFormatter).DisableColors = false // will be ignored if output is file
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

func openFile(path string) (*os.File, error) {
	// check if log directory exist, if not create one.
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(path), 0744); err != nil {
			return nil, err
		}
	}
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

// set log to specified file with specified file descriptor.
func set(logPath string, fd int) {
	logFile, err := openFile(logPath)
	if err != nil {
		log.Panicf("log: failed to open file: %v", err)
	}
	defer logFile.Close()
	if err = Dup2File(logFile, fd); err != nil {
		log.Panicf("log: failed dup 2 to path %s with error: %v", logPath, err)
	}
}

// Set to set stdout and/or stderr to a file.
// If you just want to set the stdout and stderr location without using further of these package, e.g. using your own logger.
func Set(stdOutFilePath string, stdErrFilePath string) {
	if stdOutFilePath != "" {
		set(stdOutFilePath, int(os.Stdout.Fd()))
	}
	if stdErrFilePath != "" {
		set(stdErrFilePath, int(os.Stderr.Fd()))
	}
}

// LoggerOpts options for creating new logger object.
type LoggerOpts struct {
	LogFilePath         string
	EnableTimestamp     bool
	TimestampFormat     string
	EnableRuntimeCaller bool
	IsDebug             bool
}

// New create logger object with specific location output using logrus as logger.
// If you want to specify additional logger with its own location.
func New(opts *LoggerOpts) (*_logrus.Logger, error) {
	logger := _logrus.New()

	logger.SetLevel(_logrus.InfoLevel)
	if opts.IsDebug {
		logger.SetLevel(_logrus.DebugLevel)
	}
	if opts.EnableRuntimeCaller {
		logger.ReportCaller = true
	}

	logger.Formatter = new(_logrus.TextFormatter)
	logger.Formatter.(*_logrus.TextFormatter).DisableTimestamp = true
	if opts.EnableTimestamp {
		logger.Formatter.(*_logrus.TextFormatter).DisableTimestamp = false
	}
	logger.Formatter.(*_logrus.TextFormatter).TimestampFormat = time.RFC3339
	if opts.TimestampFormat != "" {
		logger.Formatter.(*_logrus.TextFormatter).TimestampFormat = opts.TimestampFormat
	}

	logFile, err := openFile(opts.LogFilePath)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()

	logger.SetOutput(logFile)

	return logger, nil
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

func Panic(err error, msg ...interface{}) {
	logger.WithFields(_logrus.Fields{"err": err}).Panic(msg...)
}
