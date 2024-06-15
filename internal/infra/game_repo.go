package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/redis/go-redis/v9"
)

type userModel struct {
	ID        string         `redis:"id"`
	Name      string         `redis:"name"`
	Life      int            `redis:"life"`
	Sequences map[string]int `redis:"sequences"`
	DeadAt    int            `redis:"dead_at"`
	Difficult int            `redis:"difficult"`
}

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
	var game model.Game
	if err := g.redis.HGetAll(ctx, RedisGameKey).Scan(&game); err != nil {
		return nil, err
	}

	return &game, nil
}

// UpdateGame implements repository.GameRepository.
func (g *gameRepository) UpdateGame(ctx context.Context, game *model.Game) error {
	return g.redis.HSet(ctx, RedisGameKey, game.ID, game).Err()
}

// GetUserByID implements repository.GameRepository.
func (g *gameRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u userModel
	if err := g.redis.HGetAll(ctx, id).Scan(&u); err != nil {
		return nil, err
	}

	sequences := make([]*model.Sequence, 0, len(u.Sequences))

	for seq, level := range u.Sequences {
		sequences = append(sequences, &model.Sequence{
			Value: seq,
			Level: level,
		})

	}

	user := &model.User{
		ID:          u.ID,
		DisplayName: u.Name,
		Life:        u.Life,
		Sequences:   sequences,
		DeadAt:      u.DeadAt,
		Difficult:   u.Difficult,
	}

	return user, nil
}

func (g *gameRepository) UpdateUser(ctx context.Context, user *model.User) error {
	sequences := make(map[string]int)
	for _, seq := range user.Sequences {
		sequences[seq.Value] = seq.Level
	}

	u := &userModel{
		ID:        user.ID,
		Name:      user.DisplayName,
		Life:      user.Life,
		Sequences: sequences,
		DeadAt:    user.DeadAt,
		Difficult: user.Difficult,
	}

	return g.redis.HSet(ctx, user.ID, u).Err()
}
