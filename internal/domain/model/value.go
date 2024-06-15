package model

import "errors"

type RoomStatus string
type GameStatus string

const (
	GameStatusPending  GameStatus = "pending"
	GameStatusPlaying  GameStatus = "playing"
	GameStatusFinished GameStatus = "finish"
)

const (
	RoomStatusPending = "pending"
	RoomStatusMatched = "matched"
	RoomStatusPlaying = "playing"
	RoomStatusFinish  = "finish"
)

const GameStartDelay = 5 // sec
const InitUserLife = 5

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
