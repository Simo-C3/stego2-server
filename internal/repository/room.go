package repository

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain"
	"github.com/Simo-C3/stego2-server/pkg/database"
	"github.com/uptrace/bun"
)

type RoomModel struct {
	bun.BaseModel `bun:"table:rooms"`

	ID         string `bun:",pk"` // Primary Key
	Name       string `bun:"name"`
	HostName   string `bun:"host_name"`
	MinUserNum int    `bun:"min_user_num"`
	MaxUserNum int    `bun:"max_user_num"`
	UseCpu     bool   `bun:"use_cpu"`
	Status     string `bun:"status"`
}

type RoomRepository struct {
	db *database.DB
}

func convertToDomainModel(room *RoomModel) *domain.Room {
	return &domain.Room{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCpu:     room.UseCpu,
		Status:     room.Status,
	}
}

func convertToDBModel(room *domain.Room) *RoomModel {
	return &RoomModel{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCpu:     room.UseCpu,
		Status:     room.Status,
	}
}

func NewRoomRepository(db *database.DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) GetRooms(ctx context.Context) ([]*domain.Room, error) {
	var roomModels []*RoomModel
	err := r.db.NewSelect().Model(&roomModels).Scan(ctx)
	if err != nil {
		return nil, err
	}

	rooms := make([]*domain.Room, 0, len(roomModels))
	for _, roomModel := range roomModels {
		rooms = append(rooms, convertToDomainModel(roomModel))
	}

	return rooms, nil
}

func (r *RoomRepository) CreateRoom(ctx context.Context, room *domain.Room) (string, error) {
	roomModel := convertToDBModel(room)
	_, err := r.db.NewInsert().Model(roomModel).Exec(ctx)
	if err != nil {
		return "", err
	}

	return roomModel.ID, nil
}

func (r *RoomRepository) Matching(ctx context.Context) (string, error) {
	var randomRoom RoomModel
	query := r.db.NewSelect().Model(&randomRoom).OrderExpr("RAND()").Limit(1)
	err := query.Scan(ctx)
	if err != nil {
		return "", err
	}

	return randomRoom.ID, nil
}
