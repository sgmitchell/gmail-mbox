package db

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sgmitchell/gmail-mbox/gmail"
)

//go:embed schema.sql
var setupStmt []byte

type DB struct {
	*sql.DB
}

// NewDB creates a new.
func NewDB(filepath string) (*DB, error) {
	// https://avi.im/blag/2021/fast-sqlite-inserts/
	connStr := fmt.Sprintf("%s?_journal=OFF&_sync=0&_cache_size=1000000&_locking=EXCLUSIVE", filepath)
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec(string(setupStmt))
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to setup tables. %w", err)
	}
	return &DB{DB: conn}, nil
}

const (
	insMessages = `INSERT OR IGNORE INTO messages(MessageID, ThreadID,
                               FromName, FromEmail, ToLine, 
                               Subject, SizeBytes, Parts) VALUES(?,?,?,?,?,?,?,?)`
	insBodies = `INSERT OR IGNORE INTO bodies(MessageID, PlainText, HTML) VALUES(?,?,?)`
	insLabels = `INSERT OR IGNORE INTO labels(MessageID, Label) VALUES(?,?)`
	insParts  = `INSERT OR IGNORE INTO parts(MessageID, PartIdx, Headers, ContentSize) VALUES(?,?,?,?)`
)

func (db *DB) Insert(msg *gmail.Message) error {
	_, err := db.Exec(insMessages, msg.MessageID, msg.ThreadID,
		msg.From.Name, msg.From.Address, msg.To,
		msg.Subject, msg.Size, len(msg.Parts),
	)
	if err != nil {
		return err
	}

	plain, err := msg.PlainTextBody()
	if err != nil {
		return fmt.Errorf("failed to read plain text body. %w", err)
	}
	html, err := msg.HTMLBody()
	if err != nil {
		return fmt.Errorf("failed to read HTML body. %w", err)
	}

	_, err = db.Exec(insBodies, msg.MessageID, plain, html)
	if err != nil {
		return err
	}

	for _, l := range msg.Labels {
		if _, err = db.Exec(insLabels, msg.MessageID, l); err != nil {
			return err
		}
	}

	for i, p := range msg.Parts {
		hb, err := json.Marshal(p.Headers)
		if err != nil {
			return fmt.Errorf("failed to marshal headers. %w", err)
		}
		if _, err = db.Exec(insParts, msg.MessageID, i, string(hb), len(p.Body)); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Count() int {
	r, err := db.Query("SELECT COUNT(*) FROM messages")
	if err != nil || !r.Next() {
		return -1
	}
	var out int
	if err = r.Scan(&out); err != nil {
		return -1
	}
	return out
}
