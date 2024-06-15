package router

import (
	"github.com/labstack/echo/v4"

	"github.com/Simo-C3/stego2-server/internal/handler"
	myMiddleware "github.com/Simo-C3/stego2-server/pkg/middleware"
)

func InitOTPRouter(g *echo.Group, otpHandler *handler.OTPHandler, am myMiddleware.AuthController) {
	otp := g.Group("/otp")
	otp.POST("", otpHandler.GenerateOTP, am.WithHeader)
}
