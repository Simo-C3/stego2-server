//go:build wireinject

package provider

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/internal/router"
	"github.com/Simo-C3/stego2-server/internal/usecase"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/Simo-C3/stego2-server/pkg/middleware"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
)

func New(context.Context) (*echo.Echo, error) {
	wire.Build(
		// config
		config.NewDBConfig,
		config.NewRedisConfig,
		config.NewFirebaseConfig,

		// dependencies for outside
		database.New,
		redis.New,

		// repositories
		infra.NewRoomRepository,
		infra.NewGameRepository,
		infra.NewProblemRepository,
		infra.NewOTPRepository,
		infra.NewMsgSender,
		infra.NewPublisher,
		infra.NewSubscriber,

		// usecases
		usecase.NewGameManager,

		// handlers
		handler.NewRoomHandler,
		handler.NewDebugHandler,
		handler.NewOTPHandler,
		handler.NewWSHandler,
		middleware.NewAuthController,

		wire.Bind(new(service.MessageSender), new(*infra.MsgSender)),

		router.New,
	)

	return nil, nil
}
