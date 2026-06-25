package sse

import (
	"fmt"
	"sync"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type LogChan struct {
	mu      sync.Mutex
	LogsMap map[int64]chan domain.LogEvent
}

func NewSSE() *LogChan {
	return &LogChan{
		LogsMap: make(map[int64]chan domain.LogEvent),
	}
}

func (u *LogChan) Notify(deployID int64, event domain.LogEvent) {
	u.mu.Lock()
	ch, ok := u.LogsMap[deployID]
	if !ok {
		return
	}
	u.mu.Unlock()
	select {
	case ch <- event:
	default:
		fmt.Printf("warn: log channel full for deploy %s, dropping line\n", deployID)
	}
}

func (u *LogChan) AddUser(deployID int64) chan domain.LogEvent {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch, ok := u.LogsMap[deployID]
	if ok {
		return ch
	}
	// buffered chan for storing logs per user to stream via sse
	newCh := make(chan domain.LogEvent, 512)
	u.LogsMap[deployID] = newCh
	return newCh
}

func (u *LogChan) Done(deployID int64) {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch, ok := u.LogsMap[deployID]
	if !ok {
		return
	}
	close(ch)
	delete(u.LogsMap, deployID)
}
