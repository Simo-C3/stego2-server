package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/redis/go-redis/v9"
)

type OTPRepository struct {
	redis *redis.Client
}

func NewOTPRepository(redis *redis.Client) *OTPRepository {
	return &OTPRepository{
		redis: redis,
	}
}

func (r *OTPRepository) GenerateOTP(ctx context.Context, userID string) (*model.OTP, error) {
	otp, err := model.NewOTP()
	if err != nil {
		return nil, err
	}

	err = r.redis.Set(ctx, otp.OTP, userID, 0).Err()
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *OTPRepository) VerifyOTP(ctx context.Context, otp string) error {
	err := r.redis.Get(ctx, otp).Err()
	if err != nil {
		return err
	}

	err = r.redis.Del(ctx, otp).Err()
	if err != nil {
		return err
	}

	return nil
}
