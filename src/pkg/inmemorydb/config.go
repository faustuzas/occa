package inmemorydb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Configuration struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (c Configuration) Build() (Store, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Username: c.User,
		Password: c.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return RedisClient{}, fmt.Errorf("connecting to redis: %w", err)
	}

	return RedisClient{c: redisClient}, nil
}
