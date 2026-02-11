package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/anggakrnwn/kasir-api/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	var cmd string
	flag.StringVar(&cmd, "cmd", "up", "Migration command: up, reset, fresh")
	flag.Parse()

	// get database connection from env
	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		connStr = "postgresql://kasir:kasir123@localhost:5433/kasir_db?sslmode=disable"
		log.Println("Using default development database")
	}

	// connect db
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()

	// test
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping:", err)
	}

	// create migrator
	migrator := migrations.NewMigrator(db)

	// run command
	switch cmd {
	case "up":
		log.Println("running migrations...")
		if err := migrator.Up(); err != nil {
			log.Fatal("migration failed:", err)
		}
		log.Println("all migrations completed!")

	case "reset":
		log.Println("resetting database...")
		if err := migrator.Reset(); err != nil {
			log.Fatal("reset failed:", err)
		}
		log.Println("database reset!")

	case "fresh":
		log.Println("fresh migration...")
		migrator.Reset()
		if err := migrator.Up(); err != nil {
			log.Fatal("migration failed:", err)
		}
		log.Println("fresh migration completed!")

	default:
		log.Fatal("unknown command:", cmd)
	}
}
