package services

import (
	"context"
	
	"github.com/uddinArsalan/devdeploy/internals/adapters/cache"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/sse"
	"github.com/uddinArsalan/devdeploy/internals/sse/observer"
)

type LogService struct {
	cache     cache.Cache
	sse       *sse.LogChan
	observers []observer.Observer
}

func NewLogService(cache cache.Cache, sse *sse.LogChan, observers []observer.Observer) *LogService {
	return &LogService{cache: cache, sse: sse, observers: observers}
}

func (ls *LogService) StreamLogs(ctx context.Context, lastID string, deployID int64) <-chan domain.LogEvent {
	ch := ls.sse.AddUser(deployID)
	go func() {
		// this runs the continuous XRead loop and notifies via observer
		ls.cache.ReadEntriesFromStream(ctx, lastID, deployID, ls.observers)
		ls.sse.Done(deployID)
	}()
	return ch
}
