package sqlite

import (
	"path/filepath"
	"testing"
)

// =============================================================================
// InitDB
//
// InitDB has three jobs: open the connection, enable WAL + foreign keys, and
// verify the connection with a Ping. It must NOT create any schema — that is
// the seed/migration tool's responsibility.
// =============================================================================

func TestInitDB_InMemory_Succeeds(t *testing.T) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB(:memory:): %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping after InitDB: %v", err)
	}
}

func TestInitDB_FileBased_Succeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "init.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB(%q): %v", path, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping after InitDB: %v", err)
	}
}

func TestInitDB_EnablesForeignKeys(t *testing.T) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	var on int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&on); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if on != 1 {
		t.Errorf("foreign_keys = %d, want 1 (enabled)", on)
	}
}

func TestInitDB_EnablesWALOnFileDatabase(t *testing.T) {
	// WAL only applies to file-backed databases; an in-memory DB cannot use
	// WAL and SQLite will silently fall back to "memory" mode.
	path := filepath.Join(t.TempDir(), "wal.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}
}

func TestInitDB_DoesNotCreateSchema(t *testing.T) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	// InitDB must NOT provision any tables — that is the seed/migration
	// tool's job. A bare InitDB should give us an empty sqlite_master.
	var count int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master
		 WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if count != 0 {
		t.Errorf("found %d tables after InitDB; expected 0 (schema is owned by cmd/seed)", count)
	}
}

func TestInitDB_BadDataSourceName_Errors(t *testing.T) {
	// Asking SQLite to open a path inside a nonexistent directory must fail.
	// SQLite will normally create a missing file, but it cannot create a
	// missing parent directory, so this gives us a deterministic error path.
	db, err := InitDB(filepath.Join(t.TempDir(), "no-such-dir", "x.db"))
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("expected InitDB to fail when parent directory does not exist")
	}
}
