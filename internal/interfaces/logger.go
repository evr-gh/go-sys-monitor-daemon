package interfaces

import (
	"io"
)

type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
)

type Logger interface {
	io.Writer
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
	Log(level LogLevel, template string, a ...any)
}
