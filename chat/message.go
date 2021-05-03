package chat

// Message is application level message
type Message struct {
	From    string
	User    string
	Message string
}

func (m Message) IsPrivate() bool {
	return m.User != ""
}
