package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type Config struct {
	Level   string
	Out     io.Writer
	Prefix  string
	TimeFmt string
}

var (
	mu       sync.RWMutex
	minLevel           = InfoLevel
	out      io.Writer = os.Stdout
	prefix   string
	timeFmt  string
	cwd      string
)

// Initialise Logger
func Init(cfg Config) {
	mu.Lock()
	defer mu.Unlock()

	switch strings.ToLower(cfg.Level) {
	case "debug":
		minLevel = DebugLevel
	case "info":
		minLevel = InfoLevel
	case "warn", "warning":
		minLevel = WarnLevel
	case "error":
		minLevel = ErrorLevel
	case "fatal":
		minLevel = FatalLevel
	default:
		minLevel = InfoLevel
	}

	if cfg.Out != nil {
		out = cfg.Out
	} else {
		out = os.Stdout
	}

	prefix = cfg.Prefix
	timeFmt = cfg.TimeFmt

	if w, err := os.Getwd(); err == nil {
		cwd = w
	}
}

// check if the log level should be logged
func shouldLog(l Level) bool {
	mu.RLock()
	defer mu.RUnlock()
	return l >= minLevel
}

// convert level to string
func levelString(l Level) string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "INFO"
	}
}

// get caller file and line number
func callerFile(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	rel := file
	if cwd != "" {
		if r, err := filepath.Rel(cwd, file); err == nil {
			rel = r
		}
	}
	return fmt.Sprintf("%s:%d", filepath.ToSlash(rel), line)
}

// format and write the log message
func formatAndWrite(l Level, msg string) {
	if !shouldLog(l) {
		return
	}
	ts := ""
	if timeFmt != "" {
		ts = time.Now().Format(timeFmt) + " "
	}
	caller := callerFile(3) // 3 to reach the user call site (Info / Infof -> helper -> here)
	header := fmt.Sprintf("%s%s : [%s] : ", ts, levelString(l), caller)
	if prefix != "" {
		header = prefix + " " + header
	}
	// single line
	line := header + msg + "\n"
	_, _ = out.Write([]byte(line))
	if l == FatalLevel {
		os.Exit(1)
	}
}

// Info prints non-formatted info message
func Info(v ...interface{}) {
	formatAndWrite(InfoLevel, fmt.Sprint(v...))
}

// Infof prints formatted info message
func Infof(format string, v ...interface{}) {
	formatAndWrite(InfoLevel, fmt.Sprintf(format, v...))
}

// Debug prints non-formatted debug message
func Debug(v ...interface{}) {
	formatAndWrite(DebugLevel, fmt.Sprint(v...))
}

// Debugf prints formatted debug message
func Debugf(format string, v ...interface{}) {
	formatAndWrite(DebugLevel, fmt.Sprintf(format, v...))
}

// Warn prints non-formatted warning message
func Warn(v ...interface{}) {
	formatAndWrite(WarnLevel, fmt.Sprint(v...))
}

// Warnf prints formatted warning message
func Warnf(format string, v ...interface{}) {
	formatAndWrite(WarnLevel, fmt.Sprintf(format, v...))
}

// Error prints non-formatted error message
func Error(v ...interface{}) {
	formatAndWrite(ErrorLevel, fmt.Sprint(v...))
}

// Errorf prints formatted error message
func Errorf(format string, v ...interface{}) {
	formatAndWrite(ErrorLevel, fmt.Sprintf(format, v...))
}

// Fatal prints non-formatted fatal message and exits
func Fatal(v ...interface{}) {
	formatAndWrite(FatalLevel, fmt.Sprint(v...))
}

// Fatalf prints formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	formatAndWrite(FatalLevel, fmt.Sprintf(format, v...))
}
