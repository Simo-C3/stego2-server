package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type OTPRepository struct {
	redis *redis.Client
}

func NewOTPRepository(redis *redis.Client) repository.OTPRepository {
	return &OTPRepository{
		redis: redis,
	}
}

func (r *OTPRepository) GenerateOTP(ctx context.Context, userID, name string) (*model.OTP, error) {
	otp, err := model.NewOTP()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = r.redis.Set(ctx, otp.OTP, userID+";"+name, 0).Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	return otp, nil
}

func (r *OTPRepository) VerifyOTP(ctx context.Context, otp string) (string, error) {
	res, err := r.redis.Get(ctx, otp).Result()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if err := r.redis.Del(ctx, otp).Err(); err != nil {
		return "", errors.WithStack(err)
	}

	return res, nil
}
