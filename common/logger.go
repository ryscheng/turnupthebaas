package common

import (
	"io/ioutil"
	"log"
	"os"
)

var loggers = make([]*Logger, 0)
var loggersSilent = false

// Logger tracks status.
type Logger struct {
	name  string
	Trace *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

// NewLogger makes a logger.
func NewLogger(name string) *Logger {
	l := &Logger{}
	l.name = name
	l.Trace = log.New(os.Stdout, "["+name+"] TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Info = log.New(os.Stdout, "["+name+"] INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Warn = log.New(os.Stderr, "["+name+"] WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Error = log.New(os.Stderr, "["+name+"] ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	if loggersSilent {
		l.Disable()
	}
	loggers = append(loggers, l)
	return l
}

// Enable re-establishes the output for a logger
func (l *Logger) Enable() {
	l.Trace.SetOutput(os.Stdout)
	l.Info.SetOutput(os.Stdout)
	l.Warn.SetOutput(os.Stderr)
	l.Error.SetOutput(os.Stderr)
}

// Disable will stop this logger from printing
func (l *Logger) Disable() {
	l.Trace.SetOutput(ioutil.Discard)
	l.Info.SetOutput(ioutil.Discard)
	l.Warn.SetOutput(ioutil.Discard)
	l.Error.SetOutput(ioutil.Discard)
}

// SilenceLoggers will disable all loggers created with this library
func SilenceLoggers() {
	loggersSilent = true
	for _, l := range loggers {
		l.Disable()
	}
}
