package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Simo-C3/stego2-server/internal/infra"
	"github.com/Simo-C3/stego2-server/internal/usecase"
	"github.com/gorilla/websocket"
)

type Type string

const (
	TypeFinCurrentSeq Type = "FinCurrentSeq"
	TypeTypingKey     Type = "TypingKey"
)

type Base struct {
	Type    Type        `json:"type"`
	Payload interface{} `json:"payload"`
}

type FinCurrentSeq struct {
	Type    Type `json:"type"`
	Payload struct {
		Cause string `json:"cause"`
	} `json:"payload"`
}

type TypingKey struct {
	Type    Type `json:"type"`
	Payload struct {
		Key rune `json:"inputKey"`
	} `json:"payload"`
}

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

		var msg Base
		if err := json.Unmarshal(p, &msg); err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case TypeTypingKey:
			var req TypingKey
			if err := json.Unmarshal(p, &req); err != nil {
				log.Println(err)
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
			if err := h.gm.TypeKey(ctx, roomID, userID, req.Payload.Key); err != nil {
				log.Println(err)
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
		case TypeFinCurrentSeq:
			var req FinCurrentSeq
			if err := json.Unmarshal(p, &req); err != nil {
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
			h.gm.FinCurrentSeq(ctx, roomID, userID, req.Payload.Cause)
		}

		if err := ws.WriteJSON(msg); err != nil {
			break
		}
	}
}
