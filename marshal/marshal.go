package marshal

import (
	"bytes"

	"github.com/abonec/tcp_chat/chat"
	"github.com/pkg/errors"
)

var (
	ErrTagDelimiterNotFound = errors.New("tag delimiter not found in the message")
)
var (
	tagMark  = []byte("@")
	tagDelim = []byte(",")
)

// TODO: migrate to protobuf
func Marshal(m chat.Message) []byte {
	var b bytes.Buffer
	if m.User != "" {
		b.Write(tagMark)
		b.WriteString(m.User)
		b.Write(tagDelim)
	}
	b.WriteString(m.Message)
	return b.Bytes()
}

func Unmarshal(b []byte) (chat.Message, error) {
	var message chat.Message
	if bytes.HasPrefix(b, tagMark) {
		idx := bytes.Index(b, tagDelim)
		if idx == -1 {
			return message, ErrTagDelimiterNotFound
		}
		message.User = string(b[1:idx])
		b = b[idx+1:]
	}
	message.Message = string(b)
	return message, nil
}
