package cache

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type Cache interface {
	GetPort(ctx context.Context, hostname string) (int, error)
	SetHostName(ctx context.Context, hostname string, port int) error
	SetStatus(ctx context.Context, deployID int64, status domain.DeploymentStatus) error
	GetStatus(ctx context.Context, deployID int64) (domain.DeploymentStatus, error)
}
