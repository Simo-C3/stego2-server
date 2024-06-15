package model

import (
	"sort"

	"github.com/Simo-C3/stego2-server/pkg/otp"
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

type Game struct {
	ID        string
	Sequences []string
	Users     map[string]*User
	Status    GameStatus
	BaseRoom  *Room
	StartAt   int
}

type User struct {
	ID          string
	DisplayName string
	Life        int
	Sequences   []string
	Pos         int
	Streak      int
	DeadAt      int
	Difficult   int
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
		ID:        id,
		Sequences: []string{},
		Users:     map[string]*User{},
		Status:    status,
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

func (g *Game) AddUser(user *User) error {
	if g.Status != GameStatusPending {
		return ErrGameIsStarted
	}

	g.Users[user.ID] = user
	return nil
}

func (g *Game) GetRanking() []*User {
	users := make([]*User, 0, len(g.Users))
	for _, user := range g.Users {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Life > users[j].Life
	})
	// sort by life desc
	return nil
}
