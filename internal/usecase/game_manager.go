package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/Simo-C3/stego2-server/internal/schema"
)

type GameManager struct {
	pub     service.Publisher
	sub     service.Subscriber
	repo    repository.GameRepository
	problem repository.ProblemRepository
	msg     service.MessageSender
}

func NewGameManager(pub service.Publisher, sub service.Subscriber, repo repository.GameRepository, problem repository.ProblemRepository, msg service.MessageSender) *GameManager {
	return &GameManager{
		pub:     pub,
		sub:     sub,
		repo:    repo,
		problem: problem,
		msg:     msg,
	}
}

func (gm *GameManager) StartGame(ctx context.Context, roomID string) error {
	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}

	if game.ID == "" {
		fmt.Println("game is nil")
		return err
	}

	game.Status = model.GameStatusPlaying

	err = gm.repo.UpdateGame(ctx, game)
	if err != nil {
		return err
	}

	pm := &schema.PublishContent{
		RoomID: roomID,
		Payload: schema.ChangeRoomState{
			Type: schema.TypeStartGame,
			Payload: schema.ChangeRoomStatePayload{
				UserNum:    len(game.Users),
				Status:     game.Status.String(),
				StartedAt:  time.Now().Unix(),
				StartDelay: model.GameStartDelay,
				OwnerID:    game.BaseRoom.OwnerID,
			},
		},
	}

	pj, err := json.Marshal(pm)
	if err != nil {
		return err
	}

	err = gm.pub.Publish(ctx, "game", pj)
	if err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) TypeKey(ctx context.Context, gameID, userID string, key rune) error {
	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	currentSeq := user.Sequences[0]
	currentChar := currentSeq.Value[user.Pos]

	isCorrect := rune(currentChar) == key
	if isCorrect {
		user.Pos++
		user.Streak++
	} else {
		user.Streak = 0
	}

	if err := gm.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) FinCurrentSeq(ctx context.Context, roomID, userID, cause string) error {
	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}

	if cause == "succeeded" {
		// 誰かを攻撃
		// ルームから生きてるユーザーを取得
		userIDs := make([]string, 0, len(game.Users))
		for id, user := range game.Users {
			if user.Life > 0 {
				userIDs = append(userIDs, id)
			}
		}
		// ランダムに攻撃対象を選ぶ
		attackedUserIndex := rand.Intn(len(userIDs))
		// 攻撃対象のDifficultを増やす
		attackedUser, err := gm.repo.GetUserByID(ctx, userIDs[attackedUserIndex])
		if err != nil {
			return err
		}
		attackedUser.Difficult += user.Sequences[0].Level * int(math.Max(1, float64(user.Streak/10)))
		err = gm.repo.UpdateUser(ctx, attackedUser)
		if err != nil {
			return err
		}
		return nil
	} else if cause == "failed" {
		user.Life--
		if user.Life <= 0 {
			//死亡
			user.DeadAt = int(time.Now().Unix())
			err := gm.repo.UpdateUser(ctx, user)
			if err != nil {
				return err
			}
			// 順位を返す && 2位まで決まったら終了
			rank, err := game.GetRanking(userID)
			if err != nil {
				return err
			}
			result := &schema.RankResult{
				Rank: rank,
			}
			gm.msg.Send(ctx, userID, result)
			// 2位まで決まったら終了
			if rank == 2 {
				game.Status = model.GameStatusFinished
				err := gm.repo.UpdateGame(ctx, game)
				if err != nil {
					return err
				}

				// publish content
				p := &schema.PublishContent{
					RoomID: roomID,
					Payload: &schema.ChangeRoomState{
						Type: schema.TypeChangeRoom,
						Payload: schema.ChangeRoomStatePayload{
							Status: "finish",
						},
					},
				}

				publishJSON, err := json.Marshal(p)
				if err != nil {
					return err
				}
				// publish
				gm.pub.Publish(ctx, "game", publishJSON)
				return nil
			}
		}
	}

	// levelを算出
	level := user.Difficult / 100
	if level > 10 {
		level = 10
	}
	// 次の問題を取得
	problem, err := gm.problem.GetProblems(ctx, level)
	if err != nil {
		return err
	}
	nextSeq := &model.Sequence{
		Value: problem.CollectSentence,
		Level: problem.Level,
	}
	user.Sequences = append(user.Sequences[1:], nextSeq)
	user.Pos = 0

	if err := gm.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	err = gm.msg.Send(ctx, userID, map[string]interface{}{
		"type": "NextSeq",
		"payload": map[string]interface{}{
			"value": nextSeq,
			"type":  "default",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) Join(ctx context.Context, roomID, userID string) error {
	event := &schema.ChangeRoomState{
		Type: schema.TypeChangeRoom,
		Payload: schema.ChangeRoomStatePayload{
			UserNum:   1,
			Status:    model.RoomStatusMatched,
			StartedAt: time.Now().Add(30 * time.Second).Unix(),
			OwnerID:   userID,
		},
	}

	err := gm.msg.Send(ctx, userID, event)
	if err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) SubscribeMessage(ctx context.Context, topic string) {
	ch := gm.sub.Subscribe(ctx, topic)
	for msg := range ch {
		// format: roomID,payload
		fmt.Println("lets go!")
		fmt.Println("payload: ", msg.Payload)
		var content schema.PublishContent
		if err := json.Unmarshal([]byte(msg.Payload), &content); err != nil {
			fmt.Println("failed to unmarshal message:", err)
			continue
		}

		fmt.Println("roomID: ", content.RoomID)
		fmt.Println("payload: ", content.Payload)
		game, err := gm.repo.GetGameByID(ctx, content.RoomID)
		if err != nil {
			continue
		}
		userIDs := make([]string, 0, len(game.Users))
		for _, user := range game.Users {
			userIDs = append(userIDs, user.ID)
		}
		gm.msg.Broadcast(ctx, userIDs, content.Payload)
	}
}
