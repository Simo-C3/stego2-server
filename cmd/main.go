package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/Simo-C3/stego2-server/internal/repository"
	"github.com/Simo-C3/stego2-server/internal/router"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.New()
	dbCfg := config.NewDBConfig()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status}\n",
	}))

	e.GET("/", Health)

	g := e.Group("/api/v1")

	db, err := database.New(dbCfg)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	roomRepository := repository.NewRoomRepository(db)

	// Init router
	wsHandler := handler.NewWSHandler()
	roomHandler := handler.NewRoomHandler(wsHandler, roomRepository)
	router.InitRoomRouter(g, roomHandler)

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
