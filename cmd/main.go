package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/internal/router"
	"github.com/Simo-C3/stego2-server/internal/usecase"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	myMiddleware "github.com/Simo-C3/stego2-server/pkg/middleware"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.New()
	dbCfg := config.NewDBConfig()
	amCfg := config.NewFirebaseConfig()

	// middleware
	authMiddleware := myMiddleware.NewAuthController(context.Background(), amCfg)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status}\n",
	}))

	e.GET("/", Health)

	// redis疎通確認用
	// TODO: 後で消す
	e.GET("/redis-ping", func(c echo.Context) error {
		conf := config.NewRedisConfig()
		rdb, err := redis.New(conf)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		resp, err := rdb.Ping(context.Background()).Result()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		return c.JSON(http.StatusOK, resp)
	})

	g := e.Group("/api/v1")

	db, err := database.New(dbCfg)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	rdsCfg := config.NewRedisConfig()
	redis, err := redis.New(rdsCfg)
	if err != nil {
		e.Logger.Fatal(err)
	}

	roomRepository := infra.NewRoomRepository(db)
	gameRepository := infra.NewGameRepository(redis)
	publisher := infra.NewPublisher(redis)
	msgSender := infra.NewMsgSender()

	// Init router
	gm := usecase.NewGameManager(publisher, gameRepository, msgSender)
	wsHandler := handler.NewWSHandler(gm, msgSender.(*infra.MsgSender))
	roomHandler := handler.NewRoomHandler(wsHandler, roomRepository)

	// Init router
	router.InitRoomRouter(g, roomHandler, authMiddleware)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK!!👍")
}
