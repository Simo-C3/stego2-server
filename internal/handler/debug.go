package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/Simo-C3/stego2-server/pkg/redis"
	"github.com/labstack/echo/v4"
)

type DebugHandler struct {
	pub service.Publisher
}

func NewDebugHandler(pub service.Publisher) *DebugHandler {
	return &DebugHandler{
		pub: pub,
	}
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

type Event struct {
	Key     string `json:"key"`
	Payload struct {
		Name       string `json:"name"`
		HostName   string `json:"hostName"`
		MinUserNum int    `json:"minUserNum"`
		MaxUserNum int    `json:"maxUserNum"`
		UseCPU     bool   `json:"useCpu"`
		Status     string `json:"status"`
	} `json:"payload"`
}

func (h *DebugHandler) Publish(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.QueryParam("room")
	var req Event
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	p := &schema.PublishContent{
		RoomID:  id,
		Payload: req,
	}

	publishJSON, err := json.Marshal(p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.pub.Publish(ctx, "game", publishJSON); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Message Published!")
}
