package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const RedisGameKey string = "game"

type gameRepository struct {
	redis *redis.Client
}

func NewGameRepository(redis *redis.Client) repository.GameRepository {
	return &gameRepository{
		redis: redis,
	}
}

// GetGameByID implements repository.GameRepository.
func (g *gameRepository) GetGameByID(ctx context.Context, id string) (*model.Game, error) {
	data, err := g.redis.Get(ctx, id).Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var game model.Game
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, errors.WithStack(err)
	}

	return &game, nil
}

// UpdateGame implements repository.GameRepository.
func (g *gameRepository) UpdateGame(ctx context.Context, game *model.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := g.redis.Set(ctx, game.ID, data, 30*time.Minute).Err(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// GetUserByID implements repository.GameRepository.
func (g *gameRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	data, err := g.redis.Get(ctx, id).Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var user model.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, errors.WithStack(err)
	}

	return &user, nil
}

func (g *gameRepository) UpdateUser(ctx context.Context, user *model.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := g.redis.Set(ctx, user.ID, data, 30*time.Minute).Err(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeleteGame implements repository.GameRepository.
func (g *gameRepository) DeleteGame(ctx context.Context, id string) error {
	if err := g.redis.Del(ctx, id).Err(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeleteUser implements repository.GameRepository.
func (g *gameRepository) DeleteUser(ctx context.Context, id string) error {
	if err := g.redis.Del(ctx, id).Err(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
