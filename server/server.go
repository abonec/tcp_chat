package server

import (
	"context"
	"net"

	"github.com/abonec/tcp_chat/chat"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	// var buf []byte
	// msg, err := protocol.ReadMessage(conn, buf)
	// if err != nil {
	// 	reportErr(err)
	// }
	// for {
	// }
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
