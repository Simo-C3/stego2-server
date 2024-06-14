package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/Simo-C3/stego2-server/internal/router"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.New()

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

	// Init router
	wsHandler := handler.NewWSHandler()
	roomHandler := handler.NewRoomHandler(wsHandler)
	router.InitRoomRouter(g, roomHandler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := e.Start(":" + cfg.ServerPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK!!👍")
}
