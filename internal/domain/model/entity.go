package model

import (
	"sort"

	"github.com/Simo-C3/stego2-server/pkg/otp"
	"github.com/pkg/errors"
)

type Room struct {
	ID         string
	OwnerID    string
	Name       string
	HostName   string
	MinUserNum int
	MaxUserNum int
	UseCPU     bool
	Status     string
}

type Sequence struct {
	Value string
	Level int
}

type Game struct {
	ID       string
	Users    map[string]*User
	Status   GameStatus
	BaseRoom *Room
	StartAt  int
}

type User struct {
	ID          string
	DisplayName string
	Life        int
	Sequences   []*Sequence
	Pos         int
	Streak      int
	DeadAt      int
	Difficult   int
}

type Problem struct {
	ID              int
	CollectSentence string
	Level           int
}

type OTP struct {
	OTP string
}

func NewRoom(id, ownerID, name, hostName string, minUserNum, maxUserNum int, useCPU bool, status string) *Room {
	return &Room{
		ID:         id,
		OwnerID:    ownerID,
		Name:       name,
		HostName:   hostName,
		MinUserNum: minUserNum,
		MaxUserNum: maxUserNum,
		UseCPU:     useCPU,
		Status:     status,
	}
}

func NewGame(id string, status GameStatus) *Game {
	return &Game{
		ID:     id,
		Users:  map[string]*User{},
		Status: status,
	}
}

func NewUser(id, displayName string) *User {
	return &User{
		ID:          id,
		DisplayName: displayName,
		Life:        5,
	}
}

func NewOTP() (*OTP, error) {
	otp, err := otp.GenerateOTP(32)

	if err != nil {
		return nil, err
	}

	return &OTP{
		OTP: otp,
	}, nil
}

func NewProblem(id int, collectSentence string, level int) *Problem {
	return &Problem{
		ID:              id,
		CollectSentence: collectSentence,
		Level:           level,
	}
}

func (g *Game) AddUser(user *User) error {
	if g.Status != GameStatusPending {
		return ErrGameIsStarted
	}

	g.Users[user.ID] = user
	return nil
}

func (g *Game) GetRanking(userID string) (int, error) {
	deadUsers := make([]*User, 0, len(g.Users))
	for _, user := range g.Users {
		if user.Life <= 0 {
			deadUsers = append(deadUsers, user)
		}
	}

	sort.Slice(deadUsers, func(i, j int) bool {
		return deadUsers[i].DeadAt < deadUsers[j].DeadAt
	})

	for rank, user := range deadUsers {
		if user.ID == userID {
			return rank + 1, nil
		}
	}

	return -1, errors.New("user not found")
}
