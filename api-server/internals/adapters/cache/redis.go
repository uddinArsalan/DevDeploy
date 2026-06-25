package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(ctx context.Context) (*RedisClient, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        os.Getenv("REDIS_ADDR"),
		Password:    os.Getenv("REDIS_PSSWRD"),
		Username:    "default",
		DB:          0,
		ReadTimeout: -1,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Error Ping Connection Redis %v", err.Error())
	}
	return &RedisClient{
		rdb: redisClient,
	}, nil
}

func (rdb *RedisClient) SetHostName(ctx context.Context, hostname string, port int) error {
	return rdb.rdb.Set(ctx, "deploy:"+hostname+":hostname", port, 0).Err()
}

func (rdb *RedisClient) GetPort(ctx context.Context, hostname string) (int, error) {
	return rdb.rdb.Get(ctx, "deploy:"+hostname+":hostname").Int()
}

func (rdb *RedisClient) SetStatus(
    ctx context.Context,
    deployID int64,
    status domain.DeploymentStatus,
) error {
    key := fmt.Sprintf("deploy:%d:status", deployID)
    return rdb.rdb.Set(ctx, key, status, 2*time.Hour).Err()
}

func (rdb *RedisClient) GetStatus(
    ctx context.Context,
    deployID int64,
) (domain.DeploymentStatus, error) {
    key := fmt.Sprintf("deploy:%d:status", deployID)

    val, err := rdb.rdb.Get(ctx, key).Result()
    if err != nil {
        return "", err
    }

    return domain.DeploymentStatus(val), nil
}