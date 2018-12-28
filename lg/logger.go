package lg

import (
	"log"
	"os"
)

var Debug = true
var Info = true
var Err = true

func NewLogger(prefix string) *DebugLogger {
	return &DebugLogger{Log: log.New(os.Stdout, prefix, log.LstdFlags)}
}

type DebugLogger struct {
	Log *log.Logger
}

func (l *DebugLogger) IsDebugging() bool {
	return Debug
}

func (l *DebugLogger) IsInfo() bool {
	return l.IsDebugging() || Info
}
func (l *DebugLogger) IsError() bool {
	return l.IsInfo() || Err
}

func (l *DebugLogger) Debug(format string, values ...interface{}) {
	if l.IsDebugging() {
		l.Log.Printf(format, values...)
	}
}

func (l *DebugLogger) Info(format string, values ...interface{}) {
	if l.IsInfo() {
		l.Log.Printf(format, values...)
	}
}

func (l *DebugLogger) Err(format string, values ...interface{}) {
	if l.IsError() {
		l.Log.Printf(format, values...)
	}
}
