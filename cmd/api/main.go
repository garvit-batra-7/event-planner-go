package main

import (
	"database/sql"
	"log"
    _ "github.com/mattn/go-sqlite3" // SQLite driver
    "event-planner-go/internal/database"
    "event-planner-go/internal/env"
    _ "github.com/joho/godotenv/autoload"
)

type application struct {
    port int
    jwtSecret string
    models database.Models
}

func main() {
    db, err := sql.Open("sqlite3", "./data.db")
    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    models := database.NewModels(db)
    app := &application{
        port: env.GetEnvInt("PORT", 8080),
        jwtSecret: env.GetEnvString("JWT_SECRET", "secret-1234"),
        models: models,
    }

    if err := app.serve(); err != nil {
        log.Fatal(err)
    }
}
