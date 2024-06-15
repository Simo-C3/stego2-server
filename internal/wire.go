package provider

import (
	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/google/wire"
)

func New() {
	wire.Build(
		// config
		config.New,
		config.NewDBConfig,
		config.NewRedisConfig,

		// dependencies for outside
		database.New,
		redis.New,

		// repositories
		infra.NewRoomRepository,
		infra.NewGameRepository,

		// domain services

		// usecases

		// handlers

		// Pub/Sub
		infra.NewPublisher,
	)
}
