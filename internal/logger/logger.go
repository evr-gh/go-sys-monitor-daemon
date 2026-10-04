package logger

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
)

type Logger struct {
	outMu  sync.Mutex
	level  interfaces.LogLevel
	writer io.Writer
}

func New(level interfaces.LogLevel, writer io.Writer) *Logger {
	return &Logger{level: level, outMu: sync.Mutex{}, writer: writer}
}

func (l *Logger) saveMsg(level interfaces.LogLevel, template string, a ...any) {
	var buildedString strings.Builder
	fmt.Fprintf(&buildedString, "%s [%s] ", time.Now().Format("2006-01-02 15:04:05"), level)
	fmt.Fprintf(&buildedString, template, a...)
	if !strings.HasSuffix(template, "\n") {
		buildedString.WriteString("\n")
	}
	l.outMu.Lock()
	defer l.outMu.Unlock()
	_, _ = l.writer.Write([]byte(buildedString.String()))
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.outMu.Lock()
	defer l.outMu.Unlock()
	return l.writer.Write(p)
}

func (l *Logger) Debug(template string, a ...any) {
	if l.level == interfaces.DEBUG {
		l.saveMsg(interfaces.DEBUG, template, a...)
	}
}

func (l *Logger) Info(template string, a ...any) {
	if l.level == interfaces.INFO || l.level == interfaces.DEBUG {
		l.saveMsg(interfaces.INFO, template, a...)
	}
}

func (l *Logger) Warning(template string, a ...any) {
	if l.level == interfaces.WARNING || l.level == interfaces.INFO || l.level == interfaces.DEBUG {
		l.saveMsg(interfaces.WARNING, template, a...)
	}
}

func (l *Logger) Error(template string, a ...any) {
	if l.level == interfaces.ERROR || l.level == interfaces.WARNING ||
		l.level == interfaces.INFO || l.level == interfaces.DEBUG {
		l.saveMsg(interfaces.ERROR, template, a...)
	}
}

func (l *Logger) Log(level interfaces.LogLevel, template string, a ...any) {
	switch level {
	case interfaces.DEBUG:
		l.Debug(template, a...)
	case interfaces.INFO:
		l.Info(template, a...)
	case interfaces.WARNING:
		l.Warning(template, a...)
	case interfaces.ERROR:
		l.Error(template, a...)
	default:
		l.saveMsg(l.level, template, a...)
	}
}
