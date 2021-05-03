package chat

import (
	"context"
)

type User struct {
	name     string
	messages chan Message
}

func (u User) sendMessage(ctx context.Context, msg Message) {
	select {
	case <-ctx.Done():
		return
	case u.messages <- msg:
	}
}
