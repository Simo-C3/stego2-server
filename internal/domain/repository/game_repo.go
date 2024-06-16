package repository

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
)

type GameRepository interface {
	GetGameByID(ctx context.Context, id string) (*model.Game, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateGame(ctx context.Context, game *model.Game) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteGame(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error
	EditGame(ctx context.Context, gameID string, fn func(*model.Game) error) error
	EditUser(ctx context.Context, userID string, fn func(*model.User) error) error
}
