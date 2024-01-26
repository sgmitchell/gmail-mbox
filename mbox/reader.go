package mbox

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
)

const (
	// the maximum number of bytes that a single line of the mbox file might contain.
	maxLineSize = 1 * 1024 * 1024
)

// firstLinePrefix is the string matching the first line of a mbox message
var firstLinePrefix = []byte("From ")

// Reader reads an .mbox file one Message at a time.
type Reader struct {
	mu sync.Mutex
	// raw is the Reader of the underlying mbox file. It needs to be a ReadSeeker for MessageCount()
	raw io.ReadSeeker
	// scanner is the scanner used by next
	scanner *bufio.Scanner
	// lastFrom stores the last From line that marks the start of the next message
	lastFrom *bytes.Buffer
}

func NewReader(r io.ReadSeeker) *Reader {
	return &Reader{
		raw:      r,
		scanner:  bigScanner(r),
		lastFrom: bytes.NewBuffer(nil),
	}
}

// bigScanner returns a new line scanner that can accommodate longer line.
func bigScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	var b []byte
	s.Buffer(b, maxLineSize)
	return s
}

// next returns the From line and the bytes of the next message in the mbox file.
func (r *Reader) next() (string, []byte, error) {
	if r == nil {
		return "", nil, fmt.Errorf("nil reader")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	from := r.lastFrom.String()
	r.lastFrom.Reset()

	buf := bytes.NewBuffer(nil)

	for r.scanner.Scan() {
		line := r.scanner.Bytes()
		if bytes.HasPrefix(line, firstLinePrefix) {
			if len(from) == 0 {
				from = string(line)
				continue
			} else {
				r.lastFrom.Write(line)
				break
			}
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	if err := r.scanner.Err(); err != nil {
		return from, nil, err
	}
	if len(from) == 0 && buf.Len() == 0 {
		return "", nil, io.EOF
	}
	return from, buf.Bytes(), nil
}

// NextMessage does a full scan of the mbox and counts the number of message delimiters seen before seeing back to the start.
func (r *Reader) NextMessage() (*Message, error) {
	from, b, err := r.next()
	if err != nil {
		return nil, err
	}
	return NewMessage(from, b)
}

// MessageCount estimates the number of messages in the mbox.
func (r *Reader) MessageCount() (int, error) {
	if r == nil {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	returnTo, err := r.raw.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("error finding current offset. %w", err)
	}
	if _, err = r.raw.Seek(0, io.SeekStart); err != nil {
		return 0, fmt.Errorf("error seeking to beginning of file. %w", err)
	}

	s := bigScanner(r.raw)
	var numMsgs int
	for s.Scan() {
		if bytes.HasPrefix(s.Bytes(), firstLinePrefix) {
			numMsgs += 1
		}
	}

	if _, err = r.raw.Seek(returnTo, io.SeekStart); err != nil {
		return 0, fmt.Errorf("error seeking to previous position. %w", err)
	}
	return numMsgs, nil
}
