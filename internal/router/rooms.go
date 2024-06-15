package router

import (
	"github.com/labstack/echo/v4"

	"github.com/Simo-C3/stego2-server/internal/handler"
	myMiddleware "github.com/Simo-C3/stego2-server/pkg/middleware"
)

func InitRoomRouter(g *echo.Group, roomHandler *handler.RoomHandler, am myMiddleware.AuthController) {
	room := g.Group("/rooms")
	room.GET("", roomHandler.GetRooms, am.WithHeader)
	room.POST("", roomHandler.CreateRoom, am.WithHeader)
	room.GET("/matching", roomHandler.Matching, am.WithHeader)
	room.GET("/:id", roomHandler.JoinRoom)
}
