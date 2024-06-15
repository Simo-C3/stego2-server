package router

import (
	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/labstack/echo/v4"
)

func InitDebugRouter(g *echo.Group, roomHandler *handler.DebugHandler) {
	debug := g.Group("/debug")
	debug.GET("/health", roomHandler.HealthCheck)
	debug.GET("/ping-db", roomHandler.PingDB)
	debug.GET("/ping-redis", roomHandler.PingRedis)
}
