package router

import (
	"github.com/labstack/echo/v4"

	"github.com/Simo-C3/stego2-server/internal/handler"
)

func InitRoomRouter(g *echo.Group, roomHandler *handler.RoomHandler) {
	room := g.Group("/rooms")
	room.GET("", roomHandler.GetRooms)
	room.POST("", roomHandler.CreateRoom)
	room.GET("/matching", roomHandler.Matching)
}
