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
	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}

	if game.BaseRoom.OwnerID != userID {
		return errors.New("you are not owner")
	}

	game.Status = model.GameStatusPlaying

	if err = gm.repo.UpdateGame(ctx, game); err != nil {
		return err
	}

	if err = gm.roomRepo.UpdateRoom(ctx, &model.Room{
		ID:     roomID,
		Status: model.RoomStatusPlaying,
	}); err != nil {
		return err
	}

	start := time.Now().Add(10 * time.Second).Unix()
	pm := &schema.PublishContent{
		RoomID: roomID,
		Payload: schema.ChangeRoomState{
			Type: schema.TypeChangeRoom,
			Payload: schema.ChangeRoomStatePayload{
				UserNum:    len(game.Users),
				Status:     game.Status.String(),
				StartedAt:  &start,
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
	// 進捗を全体共有
	publishContent := &schema.PublishContent{
		RoomID: gameID,
		Payload: schema.TypingKey{
			Type: schema.TypeTypingKey,
			Payload: struct {
				InputSeq string "json:\"inputSeq\""
			}{
				InputSeq: key,
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
			// 攻撃対象のDifficultを増やす
			attackedUser, err := gm.repo.GetUserByID(ctx, userIDs[attackedUserIndex])
			if err != nil {
				return err
			}

			// 攻撃力を計算
			damage := user.Sequences[0].Level * int(math.Max(1, float64(user.Streak/10))) * 50
			attackedUser.Difficult += damage
			err = gm.repo.UpdateUser(ctx, attackedUser)
			if err != nil {
				return err
			}
			// gameのusersも更新
			// TODO: 排他制御
			game.Users[attackedUser.ID] = attackedUser
			if err = gm.repo.UpdateGame(ctx, game); err != nil {
				return err
			}

			// Publish: ChangeWordDifficult
			publishContent := &schema.PublishContent{
				RoomID: roomID,
				Payload: schema.Base{
					Type: schema.TypeChangeWordDifficult,
					Payload: &schema.ChangeWordDifficult{
						Difficult: attackedUser.Difficult,
						Cause:     "damage",
					},
				},
				IncludeUsers: []string{attackedUser.ID},
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
						To:     attackedUser.ID,
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
			user, err := gm.repo.GetUserByID(ctx, userID)
			if err != nil {
				return err
			}
			user.Difficult -= 200
			if user.Difficult < 0 {
				user.Difficult = 0
			}
			if err = gm.repo.UpdateUser(ctx, user); err != nil {
				return err
			}
			// gameのusersも更新
			// TODO: 排他制御
			game, err := gm.repo.GetGameByID(ctx, roomID)
			if err != nil {
				return err
			}
			game.Users[userID] = user
			if err = gm.repo.UpdateGame(ctx, game); err != nil {
				return err
			}

			event := schema.Base{
				Type: schema.TypeChangeWordDifficult,
				Payload: &schema.ChangeWordDifficult{
					Difficult: user.Difficult,
					Cause:     "heal",
				},
			}

			if err := gm.msg.Send(ctx, userID, &event); err != nil {
				return err
			}

		}
	} else if cause == "failed" {
		user, err := gm.repo.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		user.Life--
		log.Println("[251] life: ", user.Life)
		if user.Life <= 0 {
			log.Println("[253] 死亡")
			//死亡
			user.DeadAt = int(time.Now().Unix())
			err := gm.repo.UpdateUser(ctx, user)
			if err != nil {
				return err
			}
			log.Println("[260] DeadAt: ", user.DeadAt)
			// gameのusersも更新
			// TODO: 排他制御
			game, err := gm.repo.GetGameByID(ctx, roomID)
			if err != nil {
				return err
			}
			game.Users[userID] = user
			if err = gm.repo.UpdateGame(ctx, game); err != nil {
				return err
			}
			// 順位を計算
			rank, err := game.GetRanking(userID)
			if err != nil {
				return err
			}
			log.Println("[271] rank: ", rank)

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
						Rank:     rank,
					},
				},
			}
			publishJSON, err := json.Marshal(publishContent)
			if err != nil {
				return err
			}
			log.Println("[291] publishJSON: ", publishJSON)

			if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
				return err
			}

			// 2位まで決まったら終了
			if rank <= 2 {
				log.Println("[300] 終了")

				// gameのstatusを更新
				game, err := gm.repo.GetGameByID(ctx, roomID)
				if err != nil {
					return err
				}
				game.Status = model.GameStatusFinished
				if err = gm.repo.UpdateGame(ctx, game); err != nil {
					return err
				}

				// Publish: Result
				rs, err := game.GetResult()
				if err != nil {
					return err
				}
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
							OwnerID:    game.BaseRoom.OwnerID,
						},
					},
				}
				log.Println("[320] p: ", p)

				publishJSON, err = json.Marshal(p)
				if err != nil {
					return err
				}
				// publish
				if err := gm.pub.Publish(ctx, "game", publishJSON); err != nil {
					return err
				}

				log.Println("[331] game: ", game)

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
			// 生存
			if err := gm.repo.UpdateUser(ctx, user); err != nil {
				return err
			}
			// gameのusersも更新
			// TODO: 排他制御
			game, err := gm.repo.GetGameByID(ctx, roomID)
			if err != nil {
				return err
			}
			game.Users[userID] = user
			if err = gm.repo.UpdateGame(ctx, game); err != nil {
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
	level := user.Difficult / 100
	if level > 10 {
		level = 10
	}

	isHeal := rand.Intn(100) < 10
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
	user.Sequences = append(user.Sequences[1:], nextSeq)
	user.Pos = 0
	if err := gm.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	// gameのusersも更新
	// TODO: 排他制御
	game, err := gm.repo.GetGameByID(ctx, roomID)
	if err != nil {
		return err
	}
	game.Users[userID] = user
	if err = gm.repo.UpdateGame(ctx, game); err != nil {
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
			UserNum:   len(game.Users),
			Status:    game.Status.String(),
			StartedAt: nil,
			OwnerID:   game.BaseRoom.OwnerID,
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

	user, err := gm.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	problems, err := gm.problem.GetProblems(ctx, 1, 2)
	if err != nil {
		return err
	}

	for _, problem := range problems {
		user.Sequences = append(user.Sequences, &model.Sequence{
			Value: problem.CollectSentence,
			Level: problem.Level,
		})
	}

	if err = gm.repo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// gameのusersも更新
	// TODO: 排他制御
	game.Users[userID] = user
	if err = gm.repo.UpdateGame(ctx, game); err != nil {
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

	// cosp := &schema.Base{
	// 	Type: schema.TypeChangeOtherUserState,
	// 	Payload: schema.ChangeOtherUserState{
	// 		ID:       userID,
	// 		Name:     user.DisplayName,
	// 		Life:     model.InitUserLife,
	// 		Seq:      user.Sequences[0].Value,
	// 		InputSeq: "",
	// 		Rank:     0,
	// 	},
	// }

	// ev = &schema.PublishContent{
	// 	RoomID:       roomID,
	// 	Payload:      cosp,
	// 	ExcludeUsers: []string{userID},
	// }

	// marshaledCOSP, err := json.Marshal(ev)
	// if err != nil {
	// 	return err
	// }

	// if err := gm.pub.Publish(ctx, "game", marshaledCOSP); err != nil {
	// 	return err
	// }

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
