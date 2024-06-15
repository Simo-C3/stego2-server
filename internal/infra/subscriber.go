package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	redis *redis.Client
}

func NewSubscriber(redis *redis.Client) service.Subscriber {
	return &Subscriber{
		redis: redis,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string) <-chan *redis.Message {
	return s.redis.Subscribe(ctx, topic).Channel()
}
