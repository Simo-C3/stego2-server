package infra

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Simo-C3/stego2-server/internal/domain/service"
	"github.com/gorilla/websocket"
)

const writeWait = 10 * time.Millisecond

type Client struct {
	conn   *websocket.Conn
	cancel chan struct{}
	ch     chan interface{}
	err    chan error
}

func (c *Client) run() {
	for {
		select {
		case <-c.cancel:
			return
		case msg := <-c.ch:
			err := c.conn.WriteJSON(msg)
			if err != nil {
				c.err <- err
				return
			}
		}
	}
}

type MsgSender struct {
	mutex   *sync.RWMutex
	clients map[string]*Client
}

func NewMsgSender() service.MessageSender {
	return &MsgSender{
		mutex:   new(sync.RWMutex),
		clients: make(map[string]*Client),
	}
}

// Send implements service.MessageSender.
func (s *MsgSender) Send(ctx context.Context, to string, data interface{}) error {
	s.mutex.RLock()
	client, ok := s.clients[to]
	s.mutex.RUnlock()
	if !ok {
		return errors.New("client not found")
	}

	select {
	case client.ch <- data:
		return nil
	case <-time.After(writeWait):
		return errors.New("websocket write timeout")
	}
}

func (s *MsgSender) Register(userID string, conn *websocket.Conn, err chan error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client := &Client{
		conn:   conn,
		cancel: make(chan struct{}),
		ch:     make(chan interface{}, 100),
		err:    err,
	}
	go client.run()

	s.clients[userID] = client
}

func (s *MsgSender) Unregister(userID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, ok := s.clients[userID]
	if !ok {
		return
	}

	close(client.cancel)
	delete(s.clients, userID)
}
