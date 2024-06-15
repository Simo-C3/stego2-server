package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/redis/go-redis/v9"
)

type publisher struct {
	redis *redis.Client
}

func NewPublisher(redis *redis.Client) service.Publisher {
	return &publisher{
		redis: redis,
	}
}

// Publish implements service.Publisher.
func (p *publisher) Publish(ctx context.Context, topic string, data interface{}) error {
	return p.redis.Publish(ctx, topic, data).Err()
}
