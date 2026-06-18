package sse

import "sync"

type UserChan struct {
	mu      sync.Mutex
	UserMap map[string]chan interface{}
}

func NewSSE() *UserChan {
	return &UserChan{
		UserMap: make(map[string]chan interface{}),
	}
}

func (u *UserChan) Notify(userId string,data interface{}){
	u.mu.Lock()
	ch, ok := u.UserMap[userId]
	if !ok {
		ch = make(chan interface{},10)
		u.UserMap[userId] = ch
	}
	u.mu.Unlock()
	select {
	case ch <- data:
	default:
	}
}

func (u *UserChan) AddUser(userId string) chan interface{} {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch, ok := u.UserMap[userId]
	if ok {
		return ch
	}
	// buffered chan for storing logs per user to stream via sse
	newCh := make(chan interface{}, 10)
	u.UserMap[userId] = newCh
	return newCh
}

func (u *UserChan) RemoveUser(userId string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.UserMap, userId)
}


