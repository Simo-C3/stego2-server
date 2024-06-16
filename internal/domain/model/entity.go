package model

import (
	"bytes"
	"encoding/gob"
	"log"
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
	Type  string
}

type Game struct {
	ID       string
	Users    map[string]*User
	Status   GameStatus
	BaseRoom *Room
	StartAt  int
}

func (g *Game) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(g)
	return b.Bytes(), err
}

func (g *Game) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(g)
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

type GameResult struct {
	UserID      string
	DisplayName string
	Rank        int
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

func NewGame(id string, status GameStatus, room *Room) *Game {
	return &Game{
		ID:       id,
		Users:    map[string]*User{},
		Status:   status,
		BaseRoom: room,
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
		if func() bool {
			for _, u := range g.Users {
				if u.ID == user.ID {
					return true
				}
			}
			return false
		}() {
			return nil
		}
		return ErrGameIsStarted
	}

	g.Users[user.ID] = user
	return nil
}

func (g *Game) GetRanking(userID string) (int, error) {
	log.Println("GetRanking")
	deadUsers := make([]*User, 0, len(g.Users))
	for _, user := range g.Users {
		log.Println("[145] user: ", user)
		if user.Life <= 0 {
			log.Println("[147] user.Life: ", user.Life)
			deadUsers = append(deadUsers, user)
		}
	}

	num := len(g.Users)
	log.Println("[153] num: ", num)

	sort.Slice(deadUsers, func(i, j int) bool {
		return deadUsers[i].DeadAt > deadUsers[j].DeadAt
	})
	log.Println("[158] deadUsers: ", deadUsers)

	for rank, user := range deadUsers {
		if user.ID == userID {
			return num - rank, nil
		}
	}

	return 1000, errors.New("user not found")
}

func (g *Game) GetResult() ([]*GameResult, error) {
	users := make([]*User, 0, len(g.Users))
	for _, user := range g.Users {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		// DeadAtが初期値のものは1位 (DeadAtが大きいほど上位)
		if users[i].DeadAt == 0 {
			return true
		}
		if users[j].DeadAt == 0 {
			return false
		}
		return users[i].DeadAt > users[j].DeadAt
	})

	res := make([]*GameResult, 0, len(users))
	for i, user := range users {
		res = append(res, &GameResult{
			UserID:      user.ID,
			DisplayName: user.DisplayName,
			Rank:        i + 1,
		})
	}

	return res, nil
}
