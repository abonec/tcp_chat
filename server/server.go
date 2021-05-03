package server

import (
	"context"
	"net"

	"github.com/abonec/tcp_chat/chat"
	"github.com/abonec/tcp_chat/marshal"
	netclose "github.com/abonec/tcp_chat/net_closed"
	"github.com/abonec/tcp_chat/protocol"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	chat     *chat.Chat
	listener net.Listener
}

func NewServer(addr string, chat *chat.Chat) (*Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "listen for the server")
	}
	return &Server{
		listener: l,
		chat:     chat,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	defer netclose.CloseListener(s.listener)
	go func() {
		<-ctx.Done()
		netclose.CloseListener(s.listener)
	}()
	for {
		conn, err := s.listener.Accept()
		if netclose.CheckClosedError(err) {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "accept new tcp connection")
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		netclose.CloseConnection(conn)
	}()
	go func() {
		<-ctx.Done()
		netclose.CloseConnection(conn)
	}()
	usernameMessage, err := protocol.ReadMessage(conn)
	if err != nil {
		reportErr(err)
	}
	username := string(usernameMessage)
	messages := s.chat.Join(username)
	defer func() {
		s.chat.Leave(username)
	}()
	g, ctx := errgroup.WithContext(ctx)
	// handle incoming messages
	g.Go(func() error {
		for {
			protoMessage, err := protocol.ReadMessage(conn)
			if netclose.CheckClosedError(err) {
				return nil
			}
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
	// handle outgoing messages
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-messages:
				err := protocol.WriteMessage(conn, marshal.Marshal(msg))
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
	log.Err(err).Msg("got error")
}
