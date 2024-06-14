package main

import (
	"fmt"
	"net/http"

	"github.com/Simo-C3/stego2-server/pkg/config"
	"github.com/labstack/echo"
)

func main() {
	cfg := config.New()

	e := echo.New()

	e.GET("/health", Health)

	fmt.Println("Server is running on port:", cfg.ServerPort)

	e.Logger.Fatal(e.Start(":" + cfg.ServerPort))
}

func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK!!👍")
}
