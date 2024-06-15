package service

import "context"

type MessageSender interface {
	Send(ctx context.Context, to string, data interface{}) error
	Broadcast(ctx context.Context, ids []string, data interface{}) error
}
