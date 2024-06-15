package service

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Subscriber interface {
	Subscribe(ctx context.Context, topic string) <-chan *redis.Message
}
