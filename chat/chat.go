package chat

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrUserIsAlreadyConnected = errors.New("user is already connected")
	ErrUserNotFound           = errors.New("user not found")
)

type Chat struct {
	usersMu sync.RWMutex
	users   map[string]User
}

func NewChat() *Chat {
	return &Chat{
		users: make(map[string]User),
	}
}

func (c *Chat) Join(name string) <-chan Message {
	c.usersMu.Lock()
	defer c.usersMu.Unlock()
	usersMessages := make(chan Message)
	c.users[name] = User{
		name:     name,
		messages: usersMessages,
	}
	return usersMessages
}

func (c *Chat) Leave(name string) {
	c.usersMu.Lock()
	defer c.usersMu.Unlock()
	delete(c.users, name)
}

func (c *Chat) SendMessage(ctx context.Context, msg Message) error {
	if msg.IsPrivate() {
		return c.sendPrivate(ctx, msg)
	}
	c.sendBroadcast(ctx, msg)
	return nil
}

func (c *Chat) sendPrivate(ctx context.Context, msg Message) error {
	c.usersMu.RLock()
	defer c.usersMu.RUnlock()
	user, found := c.users[msg.User]
	if !found {
		return ErrUserNotFound
	}
	user.sendMessage(ctx, msg)
	return nil
}

// FIXME: this approach not scalable and will block for slow users
func (c *Chat) sendBroadcast(ctx context.Context, msg Message) {
	c.usersMu.RLock()
	defer c.usersMu.RUnlock()
	for _, u := range c.users {
		if u.name == msg.From {
			continue
		}
		u.sendMessage(ctx, msg)
	}
}
