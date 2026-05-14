package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/salary-manager/backend/internal/repository/sqlite"
	"github.com/salary-manager/backend/seed"
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

	repos := seed.Repos{
		Country:  sqlite.NewCountryRepository(db),
		Dept:     sqlite.NewDepartmentRepository(db),
		JobTitle: sqlite.NewJobTitleRepository(db),
		Employee: sqlite.NewEmployeeRepository(db),
	}

	fmt.Println("Seeding reference data (countries, departments, job titles)...")
	start := time.Now()

	fmt.Println("Seeding 10,000 employees...")
	if err := seed.SeedAll(context.Background(), repos, 10000); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	fmt.Printf("Done! Seeded reference data + 10,000 employees in %s\n", time.Since(start))
}
