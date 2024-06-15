package handler

import (
	"net/http"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/pkg/middleware"
	"github.com/labstack/echo/v4"
)

type OTPHandler struct {
	repo repository.OTPRepository
	auth middleware.AuthController
}

func NewOTPHandler(otpRepo repository.OTPRepository, auth middleware.AuthController) *OTPHandler {
	return &OTPHandler{
		repo: otpRepo,
		auth: auth,
	}
}

func convertToSchemaOTP(otp *model.OTP) *schema.OTP {
	return &schema.OTP{
		OTP: otp.OTP,
	}
}

func (h *OTPHandler) GenerateOTP(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	user, err := h.auth.GetUser(c)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	otp, err := h.repo.GenerateOTP(c.Request().Context(), userID, user.DisplayName)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return c.JSON(http.StatusOK, convertToSchemaOTP(otp))
}
