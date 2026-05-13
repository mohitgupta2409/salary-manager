package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/salary-manager/backend/internal/repository/sqlite"
	"github.com/salary-manager/backend/internal/seed"
)

func main() {
	dbPath := "salary-manager.db"

	// Remove existing DB if present for a clean seed
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Removing existing database...")
		os.Remove(dbPath)
	}

	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewEmployeeRepository(db)

	fmt.Println("Seeding 10,000 employees...")
	start := time.Now()

	if err := seed.GenerateEmployees(repo, 10000); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Done! Seeded 10,000 employees in %s\n", elapsed)
}
