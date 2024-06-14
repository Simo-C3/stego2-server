package handler

import (
	"github.com/gorilla/websocket"
)

type WSHandler struct {
}

func NewWSHandler() *WSHandler {
	return &WSHandler{}
}

func (h *WSHandler) Handle(ws *websocket.Conn) {
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}
