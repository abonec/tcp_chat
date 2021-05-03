package chat

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestChat_SendMessage(t *testing.T) {
	ctx := context.Background()
	chat := NewChat()
	firstUser := chat.Join("first")
	secondUser := chat.Join("second")
	thirdUser := chat.Join("third")
	g := &errgroup.Group{}
	broadcastMessage := Message{From: "first", Message: "broadcast"}
	sendMessage(ctx, t, g, chat, broadcastMessage)
	expectMessage(t, g, broadcastMessage, secondUser, thirdUser)
	require.NoError(t, g.Wait())
	g = &errgroup.Group{}
	privateMessage := Message{User: "first", Message: "private"}
	sendMessage(ctx, t, g, chat, privateMessage)
	expectMessage(t, g, privateMessage, firstUser)
	require.NoError(t, g.Wait())
}

func expectMessage(t *testing.T, g *errgroup.Group, expectedMessage Message, chans ...<-chan Message) {
	t.Helper()
	for _, ch := range chans {
		func(userChannel <-chan Message) {
			g.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				select {
				case <-ctx.Done():
					t.Fatalf("test was failed by timeout")
				case receivedMessage := <-userChannel:
					assert.Equal(t, expectedMessage, receivedMessage)
				}
				return nil
			})
		}(ch)
	}
}

func sendMessage(ctx context.Context, t *testing.T, g *errgroup.Group, chat *Chat, message Message) {
	t.Helper()
	g.Go(func() error {
		return chat.SendMessage(ctx, message)
	})
}
