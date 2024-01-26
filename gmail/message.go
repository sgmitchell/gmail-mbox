package gmail

import (
	"fmt"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/sgmitchell/gmail-mbox/mbox"
)

const (
	threadIdHeader = "X-Gm-Thrid"
)

// Message is a gmail message.
type Message struct {
	ThreadID  string
	MessageID string
	Date      time.Time
	From      *mail.Address
	To        string
	Subject   string
	Labels    []string

	Size int

	Parts []*BodyPart
}

// NewFromMbox converts a mbox.Message to a gmail Message.
func NewFromMbox(in *mbox.Message) (*Message, error) {
	if in == nil || in.Message == nil {
		return nil, nil
	}
	h := in.Header

	msgId, err := MessageIdFromMboxDelim(in.FromLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message id %q. %w", in.FromLine, err)
	}

	threadIdStr := h.Get(threadIdHeader)
	threadId, err := IntStrToHexStr(threadIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse thread id from %s=%q. %w", threadIdHeader, threadIdStr, err)
	}

	d, err := h.Date()
	// if a message doesn't have the date, try and use the one in the mbox delimiter
	if err != nil {
		if parts := strings.Split(in.FromLine, "@xxx "); len(parts) == 2 {
			d, err = time.Parse(time.RubyDate, parts[1])
		}
	}
	if err != nil {
		return nil, fmt.Errorf("missing date. %w", err)
	}

	from, err := mail.ParseAddress(h.Get("From"))
	if err != nil {
		from = &mail.Address{Name: h.Get("From")}
	}

	parts, err := BodyParts(textproto.MIMEHeader(h), in.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get body parts. %w", err)
	}

	return &Message{
		MessageID: msgId,
		ThreadID:  threadId,
		Date:      d,
		From:      from,
		To:        h.Get("To"),
		Subject:   h.Get("Subject"),
		Labels:    strings.Split(h.Get("X-Gmail-Labels"), ","),
		Size:      in.OriginalSize,
		Parts:     parts,
	}, nil
}

func (m *Message) PlainTextBody() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil message")
	}
	p := FirstMatchingPart(m.Parts, func(part *BodyPart) bool {
		return part.MimeType() == "text/plain"
	})
	if p != nil {
		return p.Decode()
	}
	return nil, nil
}

func (m *Message) HTMLBody() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil message")
	}
	p := FirstMatchingPart(m.Parts, func(part *BodyPart) bool {
		return part.MimeType() == "text/html"
	})
	if p != nil {
		return p.Decode()
	}
	return nil, nil
}
