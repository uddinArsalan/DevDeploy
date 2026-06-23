package queue

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type QueueMessenger interface {
	PublishMessage(ctx context.Context, job domain.BuildJob) error
	ConsumeMessage(ctx context.Context) (domain.BuildJob, error)
}
