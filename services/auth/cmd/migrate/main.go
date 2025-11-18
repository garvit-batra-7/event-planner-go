package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration direction: 'up' or 'down'")
	}
	
	direction := os.Args[1]
	
	// Open database connection
	db, err := sql.Open("sqlite3", "./data/auth.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Create database instance for migrate
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal(err)
	}
	
	// Create file source for migrations
	fileSource, err := (&file.File{}).Open("cmd/migrate/migrations")
	if err != nil {
		log.Fatal(err)
	}
	
	// Create migrate instance
	m, err := migrate.NewWithInstance("file", fileSource, "sqlite3", driver)
	if err != nil {
		log.Fatal(err)
	}
	
	// Execute migration based on direction
	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Migrations rolled back successfully")
	default:
		log.Fatal("Invalid direction. Use 'up' or 'down'")
	}
}