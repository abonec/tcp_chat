package client

import (
	"context"
	"io"
	"net"

	"github.com/abonec/tcp_chat/chat"
	"github.com/abonec/tcp_chat/marshal"
	"github.com/abonec/tcp_chat/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	addr        string
	messagesIn  chan chat.Message
	messagesOut chan chat.Message
	username    string
	conn        net.Conn
}

func NewClient(addr string, username string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "dial tcp connection")
	}
	err = registration(conn, username)
	if checkClosed(err) {
		return nil, errors.Wrap(err, "client connection was unexpectedly closed")
	}
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, username: username, messagesIn: make(chan chat.Message), messagesOut: make(chan chat.Message)}, nil
}

// SendMessage send message to the server. Start method call required before.
func (c *Client) SendMessage(ctx context.Context, msg chat.Message) {
	select {
	case <-ctx.Done():
	case c.messagesOut <- msg:
	}
}

// IncomingMessages return chan with incoming message. Start method call required before.
func (c *Client) IncomingMessages() <-chan chat.Message {
	return c.messagesIn
}

func (c *Client) Username() string {
	return c.username
}

// Start starts internal event loop that handle connection
func (c *Client) Start(ctx context.Context) error {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			reportErr(err)
		}
	}()
	g, ctx := errgroup.WithContext(ctx)
	// listen incoming messages
	g.Go(func() error {
		var buf []byte
		for {
			protoMessage, err := protocol.ReadMessage(c.conn, buf)
			if checkClosed(err) {
				return nil
			}
			if err != nil {
				return errors.Wrap(err, "read incoming message")
			}
			msg, err := marshal.Unmarshal(protoMessage)
			if err != nil {
				return errors.Wrap(err, "unmarshal proto message")
			}
			select {
			case <-ctx.Done():
				return nil
			case c.messagesIn <- msg:
			}
		}
	})
	// send upcoming messages
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-c.messagesOut:
				err := protocol.WriteMessage(c.conn, marshal.Marshal(msg))
				if err != nil {
					return errors.Wrap(err, "write proto message to the wire")
				}
				if checkClosed(err) {
					return errors.Wrap(err, "connection was unexpectedly closed")
				}
			}
		}
	})
	return errors.WithStack(g.Wait())
}

func checkClosed(err error) bool {
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
		return true
	}
	return false
}

func registration(conn net.Conn, username string) error {
	err := protocol.WriteMessage(conn, []byte(username))
	return errors.Wrap(err, "registration")
}

func reportErr(err error) {
	log.Err(err).Msg("got client error")
}
