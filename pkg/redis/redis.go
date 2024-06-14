package redis

import (
	"github.com/Simo-C3/stego2-server/pkg/config"
	redis "github.com/redis/go-redis/v9"
)

func New(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: "", // no password sets
		DB:       0,  // use default DB
	})
	return rdb, nil
}
