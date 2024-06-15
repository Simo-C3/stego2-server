package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
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
		return nil, err
	}

	var game model.Game
	if err := json.Unmarshal(data, &game); err != nil {
		return nil, err
	}

	return &game, nil
}

// UpdateGame implements repository.GameRepository.
func (g *gameRepository) UpdateGame(ctx context.Context, game *model.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}

	return g.redis.Set(ctx, game.ID, data, 30*time.Minute).Err()
}

// GetUserByID implements repository.GameRepository.
func (g *gameRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	data, err := g.redis.Get(ctx, id).Bytes()
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (g *gameRepository) UpdateUser(ctx context.Context, user *model.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return g.redis.Set(ctx, user.ID, data, 30*time.Minute).Err()
}
