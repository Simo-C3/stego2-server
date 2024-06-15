package model

import "errors"

type RoomStatus string
type GameStatus string

const (
	GameStatusPending  GameStatus = "pending"
	GameStatusPlaying  GameStatus = "playing"
	GameStatusFinished GameStatus = "finished"
)

const (
	RoomStatusPending = "pending"
	RoomStatusMatched = "matched"
	RoomStatusPlaying = "playing"
	RoomStatusFinish  = "finish"
)

func NewGameStatus(status string) GameStatus {
	switch status {
	case "pending":
		return GameStatusPending
	case "playing":
		return GameStatusPlaying
	case "finished":
		return GameStatusFinished
	default:
		return GameStatusPending
	}
}

func (s GameStatus) String() string {
	return string(s)
}

var (
	ErrMaxUserNum    error = errors.New("max user num")
	ErrGameIsStarted error = errors.New("game is started")
)
