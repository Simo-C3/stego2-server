package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
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

	e.GET("/health", Health)
	e.GET("/db-ping", func(c echo.Context) error {
		cfg := config.NewDBConfig()
		log.Println(cfg)

		db, err := database.New(cfg)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if err := db.Ping(); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, "DB is connected!!👍")
	})

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
