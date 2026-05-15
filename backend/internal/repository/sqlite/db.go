package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens the SQLite database with WAL journaling and foreign-key
// enforcement enabled, and verifies the connection.
//
// It does NOT create or migrate any schema; the database is expected to be
// provisioned by the seed/migration tool (`cmd/seed`) before the application
// is started.
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}
