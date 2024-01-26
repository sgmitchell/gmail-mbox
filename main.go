package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/sgmitchell/gmail-mbox/gmail"
	gmaildb "github.com/sgmitchell/gmail-mbox/gmail/db"
	"github.com/sgmitchell/gmail-mbox/mbox"
)

var (
	mboxFile = flag.String("in", "Takeout/Mail/All mail Including Spam and Trash.mbox", "the mbox file to read from")
	dbFile   = flag.String("out", "out.db", "the sqlite file to dump to")
)

func fatal(msg string, err error, attrs ...any) {
	attrs = append(attrs, "err", err)
	slog.Error(msg, attrs...)
	panic(msg)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()

	f, err := os.Open(*mboxFile)
	if err != nil {
		fatal("failed to open input", err)
	}
	defer f.Close()

	db, err := gmaildb.NewDB(*dbFile)
	if err != nil {
		fatal("failed to open output db", err)
	}
	defer db.Close()

	r := mbox.NewReader(f)

	slog.Info("calculating number of messages")
	expectedCount, err := r.MessageCount()
	if err != nil {
		fatal("failed to estimate number of messages", err)
	} else {
		slog.Info("done calculating number of messages", "total", expectedCount)
	}

	progress := NewProgressLog(2*time.Second, expectedCount)
	go progress.Go(ctx)

	start := time.Now()
	for {
		msg, err := r.NextMessage()
		if err == io.EOF {
			break
		} else if err != nil {
			fatal("failed to get next message", err)
		}

		progress.Add()
		if gm, err := gmail.NewFromMbox(msg); err != nil {
			slog.Error("bad parse", "err", err, "from", msg.FromLine)
		} else if err = db.Insert(gm); err != nil {
			fatal("failed insert", err, "from", msg.FromLine)
		}
	}
	actualCount := db.Count()
	missing := expectedCount - actualCount
	slog.Info("Done!", "loaded", actualCount, "missing", missing, "duration", time.Since(start).Round(time.Millisecond))
}
