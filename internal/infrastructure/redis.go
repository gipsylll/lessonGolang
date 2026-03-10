package infrastructure

import (
	"context"
	"fmt"
	"time"

	"sushkov/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	log.Info().Str("addr", cfg.Addr()).Msg("redis connected")
	return rdb, nil
}
