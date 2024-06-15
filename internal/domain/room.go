package domain

import (
	"context"
)

const (
	RoomStatusPending = "pending"
	RoomStatusMatched = "matched"
	RoomStatusPlaying = "playing"
	RoomStatusFinish  = "finish"
)

type Room struct {
	ID         string
	Name       string
	HostName   string
	MinUserNum int
	MaxUserNum int
	UseCpu     bool
	Status     string
}

func NewRoom(id, name, hostName string, minUserNum, maxUserNum int, useCpu bool, status string) *Room {
	return &Room{
		ID:         id,
		Name:       name,
		HostName:   hostName,
		MinUserNum: minUserNum,
		MaxUserNum: maxUserNum,
		UseCpu:     useCpu,
		Status:     status,
	}
}

type RoomRepository interface {
	GetRooms(ctx context.Context) ([]*Room, error)
	CreateRoom(ctx context.Context, room *Room) (string, error)
	Matching(ctx context.Context) (string, error)
}
