package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/labstack/echo/v4"
)

type DebugHandler struct {
}

func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

func (h *DebugHandler) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *DebugHandler) PingDB(c echo.Context) error {
	cfg := config.NewDBConfig()
	db, err := database.New(cfg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()
	if err := db.PingContext(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Database Connection OK!")
}

func (h *DebugHandler) PingRedis(c echo.Context) error {
	cfg := config.NewRedisConfig()
	redis, err := redis.New(cfg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()
	if _, err := redis.Ping(ctx).Result(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Redis Connection OK!")
}
