package main

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// runMigrations applies the database schema. It is safe to run repeatedly:
// every statement uses IF NOT EXISTS, so existing tables/indexes are left
// untouched.
func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
