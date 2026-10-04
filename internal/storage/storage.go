package storage

import (
	"errors"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
)

var ErrEpmtyStorage = errors.New("empty storage")

type Storage interface {
	Push(item interface{}, timestamp time.Time)
	GetElementsSince(from time.Time) <-chan interface{}
	GetElementTimestamp(item interface{}) (time.Time, bool)
	Remove(item interface{}) bool
	GetFirstElements(int64) <-chan interface{}
	Log(l interfaces.LogLevel, text string)
	GetContentInfo() []string
}
