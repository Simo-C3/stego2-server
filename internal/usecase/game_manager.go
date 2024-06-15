package usecase

import (
	"context"

	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/domain/service"
)

type GameManager struct {
	pub  service.Publisher
	repo repository.GameRepository
	msg  service.MessageSender
}

func NewGameManager(pub service.Publisher, repo repository.GameRepository, msg service.MessageSender) *GameManager {
	return &GameManager{
		pub:  pub,
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