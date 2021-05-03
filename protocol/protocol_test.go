package protocol

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWriteMessage(t *testing.T) {
	conn := bytes.NewBuffer(nil)

	message := []byte("hello world!")
	err := WriteMessage(conn, message)
	require.NoError(t, err)

	readMessage, err := ReadMessage(conn, nil)
	require.NoError(t, err)
	assert.Equal(t, message, readMessage)

	_, err = ReadMessage(conn, nil)
	assert.ErrorIs(t, err, io.EOF)
}
