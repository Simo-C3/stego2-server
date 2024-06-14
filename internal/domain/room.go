package domain

import (
	"context"
)

type Room struct {
	ID         string
	Name       string
	HostName   string
	MinUserNum int
	MaxUserNum int
	UseCpu     bool
}

func NewRoom(id, name, hostName string, minUserNum, maxUserNum int, useCpu bool) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		HostName:   hostName,
		MinUserNum: minUserNum,
		MaxUserNum: maxUserNum,
		UseCpu:     useCpu,
	}
}

type RoomRepository interface {
	GetRooms(ctx context.Context) ([]*Room, error)
	CreateRoom(ctx context.Context, room *Room) (string, error)
	Matching(ctx context.Context) (string, error)
}
