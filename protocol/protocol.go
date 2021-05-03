package protocol

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/pkg/errors"
)

const (
	lengthPartSize = 2
)

var (
	ErrMessageTooLarge              = errors.New("message too large")
	ErrWrongNumberOfBytesWasWritten = errors.New("wrong number of bytes was written to the wire")
)

func ReadMessage(r io.Reader, buf []byte) ([]byte, error) {
	buf = resizeBuffer(buf, lengthPartSize)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, errors.Wrap(err, "read length part")
	}
	// TODO: read directly to the message slice
	messageSize := int(binary.BigEndian.Uint16(buf))
	buf = resizeBuffer(buf, messageSize)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, errors.Wrap(err, "read data part")
	}
	message := make([]byte, messageSize)
	copy(message, buf)
	return message, nil
}

func WriteMessage(w io.Writer, message []byte) error {
	if len(message) > math.MaxUint16 {
		return ErrMessageTooLarge
	}
	// TODO: write directly to the writer, without allocating additional buffer
	protoMessage := make([]byte, lengthPartSize+len(message))
	binary.BigEndian.PutUint16(protoMessage, uint16(len(message)))
	copy(protoMessage[2:], message)
	n, err := w.Write(protoMessage)
	if err != nil {
		return errors.Wrap(err, "write proto message to the wire")
	}
	if n != len(protoMessage) {
		return ErrWrongNumberOfBytesWasWritten
	}
	return nil
}

func resizeBuffer(buf []byte, size int) []byte {
	if cap(buf) >= size {
		return buf[:size]
	}
	return make([]byte, size)
}
