package infra

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
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
	UseCPU     bool   `bun:"use_cpu"`
}

type roomRepository struct {
	db *database.DB
}

func NewRoomRepository(db *database.DB) repository.RoomRepository {
	return &roomRepository{
		db: db,
	}
}

func convertToDomainModel(room *RoomModel) *model.Room {
	return &model.Room{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
	}
}

func convertToDBModel(room *model.Room) *RoomModel {
	return &RoomModel{
		ID:         room.ID,
		Name:       room.Name,
		HostName:   room.HostName,
		MinUserNum: room.MinUserNum,
		MaxUserNum: room.MaxUserNum,
		UseCPU:     room.UseCPU,
	}
}

func (r *roomRepository) GetRooms(ctx context.Context) ([]*model.Room, error) {
	var roomModels []*RoomModel
	err := r.db.NewSelect().Model(&roomModels).Scan(ctx)
	if err != nil {
		return nil, err
	}

	rooms := make([]*model.Room, 0, len(roomModels))
	for _, roomModel := range roomModels {
		rooms = append(rooms, convertToDomainModel(roomModel))
	}

	return rooms, nil
}

func (r *roomRepository) CreateRoom(ctx context.Context, room *model.Room) (string, error) {
	roomModel := convertToDBModel(room)
	_, err := r.db.NewInsert().Model(roomModel).Exec(ctx)
	if err != nil {
		return "", err
	}

	return roomModel.ID, nil
}

func (r *roomRepository) Matching(ctx context.Context) (string, error) {
	var randomRoom RoomModel
	query := r.db.NewSelect().Model(&randomRoom).OrderExpr("RAND()").Limit(1)
	err := query.Scan(ctx)
	if err != nil {
		return "", err
	}

	return randomRoom.ID, nil
}

func (r *roomRepository) GetRoomByID(ctx context.Context, roomID string) (*model.Room, error) {
	var roomModel RoomModel
	err := r.db.NewSelect().Model(&roomModel).Where("id = ?", roomID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return convertToDomainModel(&roomModel), nil
}
