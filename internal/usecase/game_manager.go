package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/Simo-C3/stego2-server/internal/schema"
)

type GameManager struct {
	pub  service.Publisher
	sub  service.Subscriber
	repo repository.GameRepository
	msg  service.MessageSender
}

func NewGameManager(pub service.Publisher, sub service.Subscriber, repo repository.GameRepository, msg service.MessageSender) *GameManager {
	return &GameManager{
		pub:  pub,
		sub:  sub,
		repo: repo,
		msg:  msg,
	}
}

func (gm *GameManager) TypeKey(ctx context.Context, gameID, userID string, key rune) error {
	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	currentSeq := user.Sequences[0]
	currentChar := currentSeq[user.Pos]

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

	if cause == "succeeded" {
		// 誰かを攻撃
	} else if cause == "failed" {
		user.Life--
		if user.Life <= 0 {
			//死亡
		}
	}

	nextSeq := "例文です"

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
	event := &schema.Base{
		Type: schema.TypeChangeRoom,
		Payload: schema.ChangeRoomState{
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
		// game, err := gm.repo.GetGameByID(ctx, roomID)
		// if err != nil {
		// 	continue
		// }
		// userIDs := make([]string, 0, len(game.Users))
		// for _, user := range game.Users {
		// 	userIDs = append(userIDs, user.ID)
		// }
		// dummy IDs
		userIDs := []string{"dummyUserID"}
		gm.msg.Broadcast(ctx, userIDs, content.Payload)
	}
}
