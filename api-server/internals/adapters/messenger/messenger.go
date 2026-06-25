package queue

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type Queue interface {
	PublishMessage(ctx context.Context, job domain.BuildJob) error
	NewConsumer(ctx context.Context) (QueueConsumer, error)
	Close(ctx context.Context) error
}

type QueueConsumer interface {
	ConsumeMessage(ctx context.Context) (domain.BuildJob, error)
}
