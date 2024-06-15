package repository

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
)

type OTPRepository interface {
	GenerateOTP(ctx context.Context, userID string) (*model.OTP, error)
	VerifyOTP(ctx context.Context, otp string) (string, error)
}
