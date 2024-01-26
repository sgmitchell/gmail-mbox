package gmail

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

// BodyPart is a part of the body. Useful since almost all messages have multiple parts.
type BodyPart struct {
	Headers textproto.MIMEHeader
	Body    []byte
}

func (p *BodyPart) MimeType() string {
	ct := p.Headers.Get("Content-Type")
	if ct == "" {
		ct = "text/plain"
	}
	mediaType, _, _ := mime.ParseMediaType(ct)
	return mediaType
}

// Decode can do some additional processing on the body part's body to return it in a more useful format. Currently,
// this only does base64 decoding.
func (p *BodyPart) Decode() ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("can't decode empty bodypart")
	}
	if strings.EqualFold(p.Headers.Get("Content-Transfer-Encoding"), "base64") {
		out, err := base64.StdEncoding.DecodeString(string(p.Body))
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 body. %w", err)
		}
		return out, nil
	}
	return p.Body, nil
}

func BodyParts(headers textproto.MIMEHeader, body io.Reader) ([]*BodyPart, error) {
	ct := headers.Get("Content-Type")
	if ct == "" {
		ct = "text/plain"
	}
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("unknown media type. %q. %w", ct, err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(body, params["boundary"])
		var out []*BodyPart
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return out, nil
			} else if err != nil {
				return nil, err
			}
			parts, err := BodyParts(p.Header, p)
			if err != nil {
				return nil, err
			}
			out = append(out, parts...)
		}
	}

	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return []*BodyPart{
		{Headers: headers, Body: b},
	}, nil
}

// FirstMatchingPart is a helper for iterating a slice of BodyParts and returning the first one that evaluates f to true.
func FirstMatchingPart(parts []*BodyPart, f func(part *BodyPart) bool) *BodyPart {
	for _, p := range parts {
		if f(p) {
			return p
		}
	}
	return nil
}
