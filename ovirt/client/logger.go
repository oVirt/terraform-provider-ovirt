package client

type Logger interface {
	Log(v ...interface{})
	Logf(format string, v ...interface{})
}
