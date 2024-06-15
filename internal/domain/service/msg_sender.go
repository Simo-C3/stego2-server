package service

import "context"

type MessageSender interface {
	Send(ctx context.Context, to string, data interface{}) error
}
