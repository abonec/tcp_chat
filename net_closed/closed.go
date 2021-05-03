package netclose

import (
	"io"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func CloseConnection(conn net.Conn) {
	err := conn.Close()
	if CheckClosedError(err) {
		return
	}
	if err != nil {
		log.Err(err).Msg("failed to close connection")
	}
}

func CloseListener(l net.Listener) {
	err := l.Close()
	if CheckClosedError(err) {
		return
	}
	if err != nil {
		log.Err(err).Msg("failed to close tcp listener")
	}
}

func CheckClosedError(err error) bool {
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
		return true
	}
	return false
}
