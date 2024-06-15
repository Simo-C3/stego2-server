package repository

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
)

type RoomRepository interface {
	GetRooms(ctx context.Context) ([]*model.Room, error)
	CreateRoom(ctx context.Context, room *model.Room) (string, error)
	Matching(ctx context.Context) (string, error)
	GetRoomByID(ctx context.Context, id string) (*model.Room, error)
}
