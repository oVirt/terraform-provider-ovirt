package ovirtclient

import (
	"log"
	"testing"
)

type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}

func NewGoLogLogger() Logger {
	return &goLogLogger{}
}

type goLogLogger struct {
}

func (g *goLogLogger) Log(v ...interface{}) {
	log.Print(v...)
}

func (g *goLogLogger) Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func NewGoTestLogger(t *testing.T) Logger {
	return &goTestLogger{t: t}
}

type goTestLogger struct {
	t *testing.T
}

func (g *goTestLogger) Log(v ...interface{}) {
	g.t.Log(v...)
}

func (g *goTestLogger) Logf(format string, v ...interface{}) {
	g.t.Logf(format, v...)
}
