package usecase

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/pkg/errors"

	"github.com/Simo-C3/stego2-server/internal/domain/model"
	"github.com/Simo-C3/stego2-server/internal/domain/repository"
	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/Simo-C3/stego2-server/internal/schema"
)

type GameManager struct {
	pub      service.Publisher
	sub      service.Subscriber
	repo     repository.GameRepository
	roomRepo repository.RoomRepository
	problem  repository.ProblemRepository
	msg      service.MessageSender
}

func NewGameManager(pub service.Publisher, sub service.Subscriber, repo repository.GameRepository, roomRepo repository.RoomRepository, problem repository.ProblemRepository, msg service.MessageSender) *GameManager {
	return &GameManager{
		pub:      pub,
		sub:      sub,
		repo:     repo,
		roomRepo: roomRepo,
		problem:  problem,
		msg:      msg,
	}
}

func (gm *GameManager) StartGame(ctx context.Context, roomID string, userID string) error {
	err := gm.repo.EditGame(ctx, roomID, func(game *model.Game) error {
		if game.BaseRoom.OwnerID != userID {
			return errors.New("you are not owner")
		}

		game.Status = model.GameStatusPlaying
		return nil
	})
	if err != nil {
		return err
	}

	if err = gm.roomRepo.UpdateRoom(ctx, &model.Room{
		ID:     roomID,
		Status: model.RoomStatusPlaying,
	}); err != nil {
		return err
	}

	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}

	start := time.Now().Add(5 * time.Second).Unix()
	pm := &schema.PublishContent{
		RoomID: roomID,
		Payload: schema.ChangeRoomState{
			Type: schema.TypeChangeRoom,
			Payload: schema.ChangeRoomStatePayload{
				UserNum:    len(game.Users),
				Status:     game.Status.String(),
				StartedAt:  &start,
				StartDelay: model.GameStartDelay,
				MaxUserNum: game.BaseRoom.MaxUserNum,
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

	// 全員に問題を配布
	status := make([]*schema.ChangeOtherUserState, 0, len(game.Users))
	for _, user := range game.Users {
		status = append(status, &schema.ChangeOtherUserState{
			ID:       user.ID,
			Name:     user.DisplayName,
			Life:     user.Life,
			Seq:      user.Sequences[0].Value,
			InputSeq: "",
			Rank:     0,
		})
	}
	publishContent := &schema.PublishContent{
		RoomID: roomID,
		Payload: schema.Base{
			Type:    schema.TypeChangeOtherUsersState,
			Payload: status,
		},
	}
	publishJSON, err := json.Marshal(publishContent)
	if err != nil {
		return err
	}
	if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) TypeKey(ctx context.Context, gameID, userID string, key string) error {
	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	// 進捗を全体共有
	publishContent := &schema.PublishContent{
		RoomID: gameID,
		Payload: schema.Base{
			Type: schema.TypeChangeOtherUserState,
			Payload: schema.ChangeOtherUserState{
				ID:       userID,
				Name:     user.DisplayName,
				Life:     user.Life,
				Seq:      user.Sequences[0].Value,
				InputSeq: key,
				Rank:     0,
			},
		},
		ExcludeUsers: []string{userID},
	}
	publishJSON, err := json.Marshal(publishContent)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) FinCurrentSeq(ctx context.Context, roomID, userID, cause string) error {
	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Life <= 0 {
		return nil
	}

	if cause == "succeeded" {
		seq := user.Sequences[0]
		if seq.Type == "default" {
			// 誰かを攻撃
			// ルームから生きてるユーザーを取得
			game, err := gm.repo.GetGameByID(ctx, roomID)
			if err != nil {
				return err
			}
			userIDs := make([]string, 0, len(game.Users))
			for id, user := range game.Users {
				// 自分以外でライフが残っているユーザーを攻撃対象にする
				if user.Life > 0 && id != userID {
					userIDs = append(userIDs, id)
				}
			}

			if len(userIDs) == 0 {
				return nil
			}

			// ランダムに攻撃対象を選ぶ
			attackedUserIndex := rand.Intn(len(userIDs))

			// 攻撃力を計算
			damage := user.Sequences[0].Level * int(math.Max(1, float64(user.Streak/10))) * 20

			// User を更新
			var newDifficult int
			targetUserID := userIDs[attackedUserIndex]
			err = gm.repo.EditUser(ctx, targetUserID, func(u *model.User) error {
				u.Difficult += damage
				newDifficult = u.Difficult
				return nil
			})
			if err != nil {
				return err
			}

			err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
				g.Users[targetUserID].Difficult += damage
				return nil
			})
			if err != nil {
				return err
			}

			// Publish: ChangeWordDifficult
			publishContent := &schema.PublishContent{
				RoomID: roomID,
				Payload: schema.Base{
					Type: schema.TypeChangeWordDifficult,
					Payload: &schema.ChangeWordDifficult{
						Difficult: newDifficult,
						Cause:     "damage",
					},
				},
				IncludeUsers: []string{targetUserID},
			}
			publishJSON, err := json.Marshal(publishContent)
			if err != nil {
				return errors.WithStack(err)
			}
			if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
				return err
			}

			// Publish: AttackEvent
			publishContent = &schema.PublishContent{
				RoomID: roomID,
				Payload: schema.Base{
					Type: schema.TypeAttack,
					Payload: &schema.AttackEvent{
						From:   userID,
						To:     targetUserID,
						Damage: damage,
					},
				},
			}
			publishJSON, err = json.Marshal(publishContent)
			if err != nil {
				return errors.WithStack(err)
			}
			if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
				return err
			}
		} else if seq.Type == "heal" {
			// 回復
			var newDifficult int
			err = gm.repo.EditUser(ctx, userID, func(u *model.User) error {
				u.Difficult -= 200
				if user.Difficult < 0 {
					user.Difficult = 0
				}
				newDifficult = u.Difficult
				return nil
			})
			if err != nil {
				return err
			}

			// gameのusersも更新
			err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
				g.Users[userID].Difficult = newDifficult
				return nil
			})
			if err != nil {
				return err
			}

			event := schema.Base{
				Type: schema.TypeChangeWordDifficult,
				Payload: &schema.ChangeWordDifficult{
					Difficult: newDifficult,
					Cause:     "heal",
				},
			}

			if err := gm.msg.Send(ctx, userID, &event); err != nil {
				return err
			}

		}
	} else if cause == "failed" {
		var user *model.User
		err = gm.repo.EditUser(ctx, userID, func(u *model.User) error {
			u.Life--
			user = u
			return nil
		})
		if err != nil {
			return err
		}

		// gameのusersも更新
		err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
			g.Users[userID].Life--
			return nil
		})
		if err != nil {
			return err
		}

		if user.Life <= 0 {
			//死亡
			deadAt := int(time.Now().Unix())
			err := gm.repo.EditUser(ctx, userID, func(u *model.User) error {
				u.DeadAt = deadAt
				return nil
			})
			if err != nil {
				return err
			}

			// gameのusersも更新
			err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
				g.Users[userID].DeadAt = deadAt
				return nil
			})
			if err != nil {
				return err
			}

			// 順位を計算
			game, err := gm.repo.GetGameByID(ctx, roomID)
			if err != nil {
				return err
			}
			rank, err := game.GetRanking(userID)
			if err != nil {
				return err
			}

			// Publish: ChangeOtherUserState
			user, err := gm.repo.GetUserByID(ctx, userID)
			if err != nil {
				return err
			}
			publishContent := &schema.PublishContent{
				RoomID: roomID,
				Payload: schema.Base{
					Type: schema.TypeChangeOtherUserState,
					Payload: &schema.ChangeOtherUserState{
						ID:       user.ID,
						Name:     user.DisplayName,
						Life:     user.Life,
						Seq:      user.Sequences[0].Value,
						InputSeq: user.Sequences[0].Value[:user.Pos],
						Rank:     rank,
					},
				},
			}
			publishJSON, err := json.Marshal(publishContent)
			if err != nil {
				return err
			}

			if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
				return err
			}

			// 2位まで決まったら終了
			if rank <= 2 {
				// gameのstatusを更新
				var rs []*model.GameResult

				err := gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
					g.Status = model.GameStatusFinished
					var err error
					rs, err = game.GetResult()
					if err != nil {
						return errors.WithStack(err)
					}

					return nil
				})
				if err != nil {
					return err
				}

				// Publish: Result
				results := make([]*schema.Result, 0, len(rs))
				for _, r := range rs {
					results = append(results, schema.NewResult(r.UserID, r.Rank, r.DisplayName))
				}
				publishContent := &schema.PublishContent{
					RoomID: roomID,
					Payload: schema.Base{
						Type:    schema.TypeResult,
						Payload: results,
					},
				}
				publishJSON, err := json.Marshal(publishContent)
				if err != nil {
					return err
				}
				if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
					return err
				}

				// publish content
				p := &schema.PublishContent{
					RoomID: roomID,
					Payload: &schema.ChangeRoomState{
						Type: schema.TypeChangeRoom,
						Payload: schema.ChangeRoomStatePayload{
							UserNum:    len(game.Users),
							Status:     model.RoomStatusFinish,
							StartedAt:  nil,
							StartDelay: model.GameStartDelay,
							MaxUserNum: game.BaseRoom.MaxUserNum,
							OwnerID:    game.BaseRoom.OwnerID,
						},
					},
				}

				publishJSON, err = json.Marshal(p)
				if err != nil {
					return err
				}
				// publish
				if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
					return err
				}

				for _, user := range game.Users {
					if err := gm.repo.DeleteUser(ctx, user.ID); err != nil {
						log.Println("failed to delete user:", err)
					}
				}
				if err := gm.repo.DeleteGame(ctx, roomID); err != nil {
					return err
				}

				return nil
			}
			return nil
		} else {
			user, err := gm.repo.GetUserByID(ctx, userID)
			if err != nil {
				return err
			}
			// Publish: ChangeOtherUserState
			publishContent := &schema.PublishContent{
				RoomID: roomID,
				Payload: schema.Base{
					Type: schema.TypeChangeOtherUserState,
					Payload: &schema.ChangeOtherUserState{
						ID:       user.ID,
						Name:     user.DisplayName,
						Life:     user.Life,
						Seq:      user.Sequences[0].Value,
						InputSeq: user.Sequences[0].Value[:user.Pos],
						Rank:     0,
					},
				},
			}
			publishJSON, err := json.Marshal(publishContent)
			if err != nil {
				return err
			}

			if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
				return err
			}
		}
	}

	// levelを算出
	user, err = gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	level := user.Difficult / 100
	if level > 10 {
		level = 10
	}

	isHeal := rand.Intn(100) < 5 // 5%
	if isHeal {
		level += 3
		if level > 10 {
			level = 10
		}
	}

	// 次の問題を取得
	problems, err := gm.problem.GetProblems(ctx, level, 1)
	if err != nil {
		return err
	}
	problem := problems[0]
	typ := "default"
	if isHeal {
		typ = "heal"
	}
	nextSeq := &model.Sequence{
		Value: problem.CollectSentence,
		Level: problem.Level,
		Type:  typ,
	}

	err = gm.repo.EditUser(ctx, userID, func(u *model.User) error {
		u.Sequences = append(u.Sequences[1:], nextSeq)
		u.Pos = 0
		return nil
	})
	if err != nil {
		return err
	}

	// gameのusersも更新
	err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
		g.Users[userID].Sequences = append(g.Users[userID].Sequences[1:], nextSeq)
		g.Users[userID].Pos = 0

		return nil
	})
	if err != nil {
		return err
	}

	// Publish: NextSeq
	publishContent := &schema.PublishContent{
		RoomID: roomID,
		Payload: schema.Base{
			Type: schema.TypeNextSeq,
			Payload: &schema.NextSeqEvent{
				Value: nextSeq.Value,
				Level: nextSeq.Level,
				Type:  typ,
			},
		},
		IncludeUsers: []string{userID},
	}
	publishJSON, err := json.Marshal(publishContent)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
		return err
	}
	return nil
}

func (gm *GameManager) Join(ctx context.Context, roomID, userID string) error {
	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}

	crsp := &schema.ChangeRoomState{
		Type: schema.TypeChangeRoom,
		Payload: schema.ChangeRoomStatePayload{
			UserNum:    len(game.Users),
			Status:     game.Status.String(),
			StartedAt:  nil,
			MaxUserNum: game.BaseRoom.MaxUserNum,
			OwnerID:    game.BaseRoom.OwnerID,
		},
	}

	ev := &schema.PublishContent{
		RoomID:  roomID,
		Payload: crsp,
	}

	marshaledCRSP, err := json.Marshal(ev)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := gm.pub.Publish(ctx, "game", marshaledCRSP); err != nil {
		return err
	}

	var user *model.User
	err = gm.repo.EditUser(ctx, userID, func(u *model.User) error {
		problems, err := gm.problem.GetProblems(ctx, 1, 2)
		if err != nil {
			return err
		}

		for _, problem := range problems {
			u.Sequences = append(u.Sequences, &model.Sequence{
				Value: problem.CollectSentence,
				Level: problem.Level,
			})
		}

		user = u

		return nil
	})
	if err != nil {
		return err
	}

	// gameのusersも更新
	err = gm.repo.EditGame(ctx, roomID, func(g *model.Game) error {
		g.Users[userID] = user
		return nil
	})
	if err != nil {
		return err
	}

	if err = gm.msg.Send(ctx, userID, &schema.Base{
		Type: schema.TypeNextSeq,
		Payload: schema.NextSeqEvent{
			Value: user.Sequences[0].Value,
			Type:  "default",
			Level: user.Sequences[0].Level,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (gm *GameManager) SubscribeMessage(ctx context.Context, topic string) {
	ch := gm.sub.Subscribe(ctx, topic)
	for msg := range ch {
		// format: roomID,payload
		var content schema.PublishContent
		if err := json.Unmarshal([]byte(msg.Payload), &content); err != nil {
			log.Println("failed to unmarshal message:", err)
			continue
		}

		game, err := gm.repo.GetGameByID(ctx, content.RoomID)
		if err != nil {
			continue
		}
		userIDs := make([]string, 0, len(game.Users))

		includeUsers := content.IncludeUsers
		excludeUsers := content.ExcludeUsers

		for _, user := range game.Users {
			if slices.Contains(excludeUsers, user.ID) {
				continue
			}

			if len(includeUsers) > 0 && !slices.Contains(includeUsers, user.ID) {
				continue
			}

			userIDs = append(userIDs, user.ID)
		}

		if err := gm.msg.Broadcast(ctx, userIDs, content.Payload); err != nil {
			log.Println("failed to broadcast message:", err)
		}
	}
}
