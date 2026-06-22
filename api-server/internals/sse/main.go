package sse

import "sync"

type LogChan struct {
	mu      sync.Mutex
	LogsMap map[string]chan interface{}
}

func NewSSE() *LogChan {
	return &LogChan{
		LogsMap: make(map[string]chan interface{}),
	}
}

func (u *LogChan) Notify(deploymentId string,data interface{}){
	u.mu.Lock()
	ch, ok := u.LogsMap[deploymentId]
	if !ok {
		ch = make(chan interface{},10)
		u.LogsMap[deploymentId] = ch
	}
	u.mu.Unlock()
	select {
	case ch <- data:
	default:
	}
}

func (u *LogChan) AddUser(deploymentId string) chan interface{} {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch, ok := u.LogsMap[deploymentId]
	if ok {
		return ch
	}
	// buffered chan for storing logs per user to stream via sse
	newCh := make(chan interface{}, 10)
	u.LogsMap[deploymentId] = newCh
	return newCh
}

func (u *LogChan) RemoveUser(deploymentId string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.LogsMap, deploymentId)
}


