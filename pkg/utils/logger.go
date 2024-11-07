package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Logger interface {
	Debug(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type LogMode int

const (
	LogModeProduction LogMode = iota
	LogModeDebug
)

// LoggerImpl implements the Logger interface
type LoggerImpl struct {
	mode LogMode
}

// NewLogger returns a new Logger implementation
func NewLogger(debug bool) Logger {
	mode := LogModeProduction
	if debug {
		mode = LogModeDebug
	}
	return &LoggerImpl{mode: mode}
}

func (l *LoggerImpl) Debug(format string, args ...interface{}) {
	if l.mode >= LogModeDebug {
		_, file, line, _ := runtime.Caller(1)
		msg := fmt.Sprintf(format, args...)
		fmt.Printf("DEBUG [%s:%d]: %s\n", filepath.Base(file), line, msg)
	}
}

func (l *LoggerImpl) Error(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, args...)
	if l.mode >= LogModeDebug {
		fmt.Fprintf(os.Stderr, "ERROR [%s:%d]: %s\n", filepath.Base(file), line, msg)
	}
}
