package mbox

import (
	"bytes"
	"net/mail"
)

// Message is a message in an mbox file. This is.
type Message struct {
	*mail.Message
	// FromLine is the delimiter line that started this message
	FromLine string
	// OriginalSize is the original size in bytes of the mbox message
	OriginalSize int
}

func NewMessage(fromLine string, raw []byte) (*Message, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return &Message{
		Message:      msg,
		FromLine:     fromLine,
		OriginalSize: len(raw),
	}, nil
}
