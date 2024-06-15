package router

import (
	"github.com/Simo-C3/stego2-server/internal/handler"
	"github.com/Simo-C3/stego2-server/pkg/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func New(
	debugHandler *handler.DebugHandler,
	roomHandler *handler.RoomHandler,
	otpHandler *handler.OTPHandler,
	amHandler middleware.AuthController,
) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.Recover())

	g := e.Group("/api/v1")
	g.Use(echoMiddleware.Logger())
	g.Use(echoMiddleware.CORS())
	g.Use(echoMiddleware.Gzip())

	InitDebugRouter(g, debugHandler)
	InitRoomRouter(g, roomHandler, amHandler)
	InitOTPRouter(g, otpHandler, amHandler)

	return e
}
