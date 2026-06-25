package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/uddinArsalan/devdeploy/internals/domain"
	"github.com/uddinArsalan/devdeploy/internals/sse/observer"
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

func (rdb *RedisClient) SetStatus(ctx context.Context, deployID int64, status domain.DeploymentStatus) error {
    key := fmt.Sprintf("deploy:%d:status", deployID)
    return rdb.rdb.Set(ctx, key, string(status), 2*time.Hour).Err()
}

func (rdb *RedisClient) GetStatus(ctx context.Context, deployID int64) (domain.DeploymentStatus, error) {
    key := fmt.Sprintf("deploy:%d:status", deployID)
    val, err := rdb.rdb.Get(ctx, key).Result()
    if err != nil {
        return "", err
    }
    return domain.DeploymentStatus(val), nil
}

func (rdb *RedisClient) AppendLogsAndStatus(ctx context.Context, logType domain.LogType, data interface{}, deployID int64) error {
	streamName := fmt.Sprintf("stream:%d", deployID)
	return rdb.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"type": string(logType),
			"data": data,
		},
		MaxLen: 2000,
		Approx: true,
	}).Err()
}

func (rdb *RedisClient) ReadEntriesFromStream(ctx context.Context, lastID string, deployID int64, observers []observer.Observer) error {
	streamName := fmt.Sprintf("stream:%d", deployID)
	cursor := lastID

	for {
		entries, err := rdb.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, cursor},
			Count:   100,
			Block:   500,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return nil // client disconnected
			}
			continue
		}

		for _, val := range entries {
			for _, msg := range val.Messages {
				logType := domain.LogType(msg.Values["type"].(string))
				data := msg.Values["data"].(string)
				cursor = msg.ID
				event := domain.LogEvent{
					ID:   msg.ID,
					Type: logType,
					Data: data,
				}
				for _, obs := range observers {
					obs.Notify(deployID, event)
				}
			}
		}

	}
}
