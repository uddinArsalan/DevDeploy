package cache

import (
	"context"

	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/sse/observer"
)

type Cache interface {
	GetPort(ctx context.Context, hostname string) (int, error)
	SetHostName(ctx context.Context, hostname string, port int) error
	SetStatus(ctx context.Context, deployID int64, status domain.DeploymentStatus) error
	GetStatus(ctx context.Context, deployID int64) (domain.DeploymentStatus, error)
	AppendLogsAndStatus(ctx context.Context, logType domain.LogType, data interface{}, deployID int64) error
	ReadEntriesFromStream(ctx context.Context, lastID string, deployID int64, obs []observer.Observer) error
}
