package handler

import (
	"context"
	"encoding/json"

	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/internal/usecase"
	"github.com/Simo-C3/stego2-server/pkg/logger"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	gm        *usecase.GameManager
	msgSender *infra.MsgSender
}

func NewWSHandler(gm *usecase.GameManager, sender *infra.MsgSender) *WSHandler {
	return &WSHandler{
		gm:        gm,
		msgSender: sender,
	}
}

func (h *WSHandler) Handle(ctx context.Context, ws *websocket.Conn, roomID, userID string) {
	errCh := make(chan error)
	defer close(errCh)

	h.msgSender.Register(userID, ws, errCh)
	defer h.msgSender.Unregister(userID)

	logger := logger.New()

	if err := h.gm.Join(ctx, roomID, userID); err != nil {
		logger.LogErrorWithStack(ctx, err)
	}

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			logger.LogErrorWithStack(ctx, err)
			break
		}

		var msg schema.Base
		if err := json.Unmarshal(p, &msg); err != nil {
			logger.LogErrorWithStack(ctx, err)
			break
		}

		switch msg.Type {
		case schema.TypeTypingKey:
			var req schema.TypingKey
			if err := json.Unmarshal(p, &req); err != nil {
				logger.LogErrorWithStack(ctx, err)
			}
			if err := h.gm.TypeKey(ctx, roomID, userID, req.Payload.InputSeq, req.Payload.UserID); err != nil {
				logger.LogErrorWithStack(ctx, err)
			}
		case schema.TypeFinCurrentSeq:
			var req schema.FinCurrentSeq
			if err := json.Unmarshal(p, &req); err != nil {
				logger.LogErrorWithStack(ctx, err)
			}
			if err := h.gm.FinCurrentSeq(ctx, roomID, userID, req.Payload.Cause); err != nil {
				logger.LogErrorWithStack(ctx, err)
			}
		case schema.TypeStartGame:
			if err := h.gm.StartGame(ctx, roomID, userID); err != nil {
				logger.LogErrorWithStack(ctx, err)
			}
		}
	}
}

func (h *WSHandler) SubscribeHandle(ctx context.Context, topic string) {
	h.gm.SubscribeMessage(ctx, topic)
}
