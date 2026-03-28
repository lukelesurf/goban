package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	// Column additions for existing databases — errors are intentionally ignored
	// (ALTER TABLE fails with "duplicate column name" if already applied)
	conn.Exec(`ALTER TABLE tasks ADD COLUMN due_date DATETIME`)
	conn.Exec(`ALTER TABLE boards ADD COLUMN is_default INTEGER NOT NULL DEFAULT 0`)

	return &DB{conn}, nil
}
