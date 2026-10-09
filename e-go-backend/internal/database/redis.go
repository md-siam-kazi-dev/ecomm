package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
)

// NewRedis parses REDIS_URL, connects, and verifies the connection with PING.
// An empty URL returns (nil, nil) so Redis remains optional in development.
func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.URL == "" {
		return nil, nil
	}

	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse REDIS_URL: %w", err)
	}

	opt.PoolSize = cfg.PoolSize
	opt.DialTimeout = cfg.DialTimeout
	opt.ReadTimeout = cfg.ReadTimeout
	opt.WriteTimeout = cfg.WriteTimeout

	client := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return client, nil
}

// PingRedis verifies Redis connectivity, used by the readiness probe.
func PingRedis(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}
