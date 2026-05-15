package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/salary-manager/backend/internal/repository/sqlite"
)

func main() {
	dbPath := "salary-manager.db"

	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Removing existing database...")
		os.Remove(dbPath)
		os.Remove(dbPath + "-shm")
		os.Remove(dbPath + "-wal")
	}

	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Pin to a single connection so the seed-only PRAGMAs below stick for
	// every query the seeder issues. database/sql may otherwise hand out
	// fresh connections that don't inherit the relaxed durability settings.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`
		PRAGMA journal_mode = MEMORY;
		PRAGMA synchronous  = OFF;
		PRAGMA temp_store   = MEMORY;
		PRAGMA foreign_keys = OFF;
	`); err != nil {
		log.Fatalf("failed to apply seed pragmas: %v", err)
	}

	start := time.Now()

	fmt.Println("Running migrations...")
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	repos := Repos{
		Country:  sqlite.NewCountryRepository(db),
		Dept:     sqlite.NewDepartmentRepository(db),
		JobTitle: sqlite.NewJobTitleRepository(db),
		Employee: sqlite.NewEmployeeRepository(db),
	}

	fmt.Println("Seeding reference data (countries, departments, job titles)...")
	fmt.Println("Seeding 10,000 employees...")
	if err := SeedAll(context.Background(), db, repos, 10000); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	if err := restoreProductionPragmas(db); err != nil {
		log.Fatalf("failed to restore production pragmas: %v", err)
	}

	fmt.Printf("Done! Migrated schema + seeded reference data + 10,000 employees in %s\n", time.Since(start))
}

// restoreProductionPragmas re-enables FK enforcement, validates that the bulk
// load did not leave any orphan rows, switches back to WAL with normal
// durability, and refreshes planner statistics. The result is a database file
// that matches what InitDB expects when the API boots.
func restoreProductionPragmas(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("re-enable foreign keys: %w", err)
	}

	rows, err := db.Query(`PRAGMA foreign_key_check;`)
	if err != nil {
		return fmt.Errorf("foreign_key_check: %w", err)
	}
	hasViolation := rows.Next()
	rows.Close()
	if hasViolation {
		return fmt.Errorf("foreign key check failed after seed")
	}

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		return fmt.Errorf("restore WAL: %w", err)
	}
	if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil {
		return fmt.Errorf("restore synchronous: %w", err)
	}
	if _, err := db.Exec(`ANALYZE;`); err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	return nil
}
