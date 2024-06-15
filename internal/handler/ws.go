package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/internal/schema"
	"github.com/Simo-C3/stego2-server/internal/usecase"
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

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var msg schema.Base
		if err := json.Unmarshal(p, &msg); err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case schema.TypeTypingKey:
			var req schema.TypingKey
			if err := json.Unmarshal(p, &req); err != nil {
				log.Println(err)
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
			if err := h.gm.TypeKey(ctx, roomID, userID, req.Payload.Key); err != nil {
				log.Println(err)
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
		case schema.TypeFinCurrentSeq:
			var req schema.FinCurrentSeq
			if err := json.Unmarshal(p, &req); err != nil {
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
			h.gm.FinCurrentSeq(ctx, roomID, userID, req.Payload.Cause)
		case schema.TypeStartGame:
			if err := h.gm.StartGame(ctx, roomID); err != nil {
				fmt.Println(err)
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
		}

		if err := ws.WriteJSON(msg); err != nil {
			break
		}
	}
}

func (h *WSHandler) SubscribeHandle(ctx context.Context, topic string) {
	h.gm.SubscribeMessage(ctx, topic)
}
