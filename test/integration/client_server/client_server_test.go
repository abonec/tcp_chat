package client_server

import (
	"context"
	"testing"
	"time"

	"github.com/abonec/tcp_chat/chat"
	"github.com/abonec/tcp_chat/client"
	"github.com/abonec/tcp_chat/server"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestClientServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	addr, stopServer := startServer(ctx, t)
	defer stopServer()
	client1, done1 := connectClient(ctx, t, addr, "first")
	client2, done2 := connectClient(ctx, t, addr, "second")
	client3, done3 := connectClient(ctx, t, addr, "third")
	broadcastMessage := chat.Message{Message: "broadcast message"}
	sendMessage(ctx, client1, broadcastMessage)
	expectMessage(ctx, t, broadcastMessage, client2, client3)

	privateMessage := chat.Message{User: "second", Message: "private message"}
	sendMessage(ctx, client2, privateMessage)
	expectMessage(ctx, t, privateMessage, client2)

	cancel()
	<-done1
	<-done2
	<-done3
}

func expectMessage(ctx context.Context, t *testing.T, expectMessage chat.Message, clients ...*client.Client) {
	g, ctx := errgroup.WithContext(ctx)
	for _, cli := range clients {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		currentClient := cli
		g.Go(func() error {
			defer cancel()
			select {
			case <-timeoutCtx.Done():
				t.Fatalf("failed to get message from %s", currentClient.Username())
			case msg := <-currentClient.IncomingMessages():
				assert.Equal(t, expectMessage, msg)
			}
			return nil
		})
	}
	require.NoError(t, g.Wait())
}

func sendMessage(ctx context.Context, cli *client.Client, message chat.Message) {
	go cli.SendMessage(ctx, message)
}

func connectClient(ctx context.Context, t *testing.T, addr string, userName string) (*client.Client, chan struct{}) {
	t.Helper()
	cli, err := client.NewClient(addr, userName)
	require.NoError(t, err)
	done := make(chan struct{})
	go func() {
		err := cli.Start(ctx)
		assert.NoError(t, err)
		close(done)
	}()
	return cli, done
}

func startServer(ctx context.Context, t *testing.T) (addr string, stop func()) {
	t.Helper()
	chatApp := chat.NewChat()
	s, err := server.NewServer(":0", chatApp)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		err := s.Start(ctx)
		if err != nil {
			reportErr(err)
		}
	}()
	return s.Addr(), cancel
}

func reportErr(err error) {
	log.Err(err).Msg("go error")
}
