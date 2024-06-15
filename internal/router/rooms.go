package router

import (
	"github.com/labstack/echo/v4"

	"github.com/Simo-C3/stego2-server/internal/handler"
	myMiddleware "github.com/Simo-C3/stego2-server/pkg/middleware"
)

func InitRoomRouter(g *echo.Group, roomHandler *handler.RoomHandler, am myMiddleware.AuthController) {
	room := g.Group("/rooms", am.WithHeader)
	room.GET("", roomHandler.GetRooms)
	room.POST("", roomHandler.CreateRoom)
	room.GET("/matching", roomHandler.Matching)
	room.GET("/:id", roomHandler.JoinRoom)
}
