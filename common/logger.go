package common

import (
	"log"
	"os"
)

type Logger struct {
	name  string
	Trace *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

func NewLogger(name string) *Logger {
	l := &Logger{}
	l.name = name
	l.Trace = log.New(os.Stdout, "["+name+":TRACE] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Info = log.New(os.Stdout, "["+name+":INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Warn = log.New(os.Stderr, "["+name+":WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Error = log.New(os.Stderr, "["+name+":ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	return l
}
