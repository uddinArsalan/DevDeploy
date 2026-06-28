package queue

import (
	"context"
	"time"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type Queue interface {
	PublishMessage(ctx context.Context, job domain.BuildJob) error
	NewConsumer(ctx context.Context) (QueueConsumer, error)
	Close(ctx context.Context) error
}

type Delivery interface {
	Ack(ctx context.Context) error
	Retry(ctx context.Context) error
	Reject(ctx context.Context,description string) error
	DelayRetry(ctx context.Context, delay time.Duration) error 
}

type QueueConsumer interface {
	ConsumeMessage(ctx context.Context) (domain.BuildJob, Delivery, error)
}
