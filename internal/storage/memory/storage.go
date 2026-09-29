package memorystorage

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/storage"
)

type element struct {
	timestamp time.Time
	data      interface{}
}

type MemoryStorage struct {
	name   string
	logger interfaces.Logger
	rwm    sync.RWMutex
	list   *list.List
	size   int64
}

func New(logger interfaces.Logger, name string, sizeLimit int64) *MemoryStorage {
	return &MemoryStorage{logger: logger, name: name, rwm: sync.RWMutex{}, list: list.New(), size: sizeLimit}
}

func (ms *MemoryStorage) Push(s interface{}, t time.Time) {
	ms.rwm.Lock()
	defer ms.rwm.Unlock()

	if ms.size == 0 {
		return
	}
	if ms.list.Len() == int(ms.size) {
		elem := ms.list.Back()
		ms.list.Remove(elem)
		ms.logger.Debug("Из хранилища %q удален элемент: %s, %+v",
			ms.name, elem.Value.(element).timestamp, elem.Value.(element).data)
	}
	ms.list.PushFront(element{timestamp: t, data: s})
	ms.logger.Debug("В хранилище %q добавлен элемент: %s, %+v", ms.name, t, s)
}

func (ms *MemoryStorage) GetElementsSince(t time.Time) <-chan interface{} {
	elemCh := make(chan interface{})
	go func() {
		ms.rwm.RLock()
		defer close(elemCh)
		defer ms.rwm.RUnlock()
		for last := ms.list.Front(); last != nil; last = last.Next() {
			elem := last.Value.(element)
			if t.After(elem.timestamp) {
				return
			}
			elemCh <- elem.data
		}
	}()

	return elemCh
}

func (ms *MemoryStorage) GetFirstElements(num int64) <-chan interface{} {
	elemCh := make(chan interface{})
	go func() {
		ms.rwm.RLock()
		defer close(elemCh)
		defer ms.rwm.RUnlock()
		last := ms.list.Front()
		for num > 0 {
			if last == nil {
				break
			}
			elem := last.Value.(element)
			elemCh <- elem.data
			last = last.Next()
			num--
		}
	}()

	return elemCh
}

func (ms *MemoryStorage) GetContentInfo() []string {
	ms.rwm.RLock()
	defer ms.rwm.RUnlock()

	res := make([]string, ms.list.Len(), 0)
	for e := ms.list.Front(); e != nil; e = e.Next() {
		res = append(res, fmt.Sprintf("%s: %+v", e.Value.(element).timestamp, e.Value.(element).data))
	}
	return res
}

func (ms *MemoryStorage) Log(l interfaces.LogLevel, text string) {
	ms.logger.Log(l, "%s : %v", text, ms.GetContentInfo())
}

func (ms *MemoryStorage) GetElementTimestamp(value interface{}) (time.Time, bool) {
	ms.rwm.RLock()
	defer ms.rwm.RUnlock()

	for e := ms.list.Front(); e != nil; e = e.Next() {
		elem := e.Value.(element)
		if elem.data == value {
			return elem.timestamp, true
		}
	}
	return time.Time{}, false
}

func (ms *MemoryStorage) Remove(value interface{}) bool {
	ms.rwm.Lock()
	defer ms.rwm.Unlock()

	for e := ms.list.Front(); e != nil; e = e.Next() {
		elem := e.Value.(element)
		if elem.data == value {
			ms.list.Remove(e)
			ms.logger.Debug("Из хранилища %q удален элемент: %+v", ms.name, elem)
			return true
		}
	}
	return false
}

func (ms *MemoryStorage) Clean(t time.Time) {
	ms.rwm.Lock()
	defer ms.rwm.Unlock()

	ms.logger.Debug("Удаление из хранилища %q старых элементов (%s)", ms.name, t)
	for e := ms.list.Back(); e != nil; {
		elem := e.Value.(element)
		if t.After(elem.timestamp) {
			next := e.Prev()
			ms.list.Remove(e)
			e = next
		} else {
			e = e.Prev()
		}
	}
}

var _ storage.Storage = (*MemoryStorage)(nil)
