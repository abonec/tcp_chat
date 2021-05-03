package marshal

import (
	"testing"

	"github.com/abonec/tcp_chat/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_Serialize(t *testing.T) {
	testMessage(t, chat.Message{User: "test", Message: "message"})
	testMessage(t, chat.Message{Message: "message"})
}

func testMessage(t *testing.T, message chat.Message) {
	t.Helper()

	deserializedMessage, err := Unmarshal(Marshal(message))
	require.NoError(t, err)
	assert.Equal(t, message, deserializedMessage)
}
