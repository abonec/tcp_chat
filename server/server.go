package server

import (
	"context"
	"net"

	"github.com/abonec/tcp_chat/chat"
	"github.com/abonec/tcp_chat/marshal"
	"github.com/abonec/tcp_chat/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	addr string
	chat *chat.Chat
}

func NewServer(addr string, chat *chat.Chat) *Server {
	return &Server{
		addr: addr,
		chat: chat,
	}
}

func (s *Server) Start(ctx context.Context) error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return errors.Wrap(err, "listen for the server")
	}
	defer closeListener(l)
	go func() {
		<-ctx.Done()
		closeListener(l)
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			return errors.Wrap(err, "accept new tcp connection")
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		closeConnection(conn)
	}()
	go func() {
		<-ctx.Done()
		closeConnection(conn)
	}()
	var buf []byte
	usernameMessage, err := protocol.ReadMessage(conn, buf)
	if err != nil {
		reportErr(err)
	}
	username := string(usernameMessage)
	messages := s.chat.Join(username)
	defer func() {
		s.chat.Leave(username)
	}()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for {
			protoMessage, err := protocol.ReadMessage(conn, buf)
			if err != nil {
				return errors.Wrap(err, "read proto message from client")
			}
			msg, err := marshal.Unmarshal(protoMessage)
			if err != nil {
				return errors.Wrap(err, "unmarshal message from client")
			}
			msg.From = username
			err = s.chat.SendMessage(ctx, msg)
			if err != nil {
				return errors.Wrap(err, "send message to the chat")
			}
		}
	})
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-messages:
				_, err := conn.Write(marshal.Marshal(msg))
				if err != nil {
					return errors.Wrap(err, "writing message to client")
				}
			}
		}
	})
	err = g.Wait()
	if err != nil {
		reportErr(err)
	}
}

func reportErr(err error) {
	log.Err(err).Msg("go error")
}

func closeConnection(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Err(err).Msg("failed to close connection")
	}
}

func closeListener(l net.Listener) {
	err := l.Close()
	if err != nil {
		log.Err(err).Msg("failed to close tcp listener")
	}
}
