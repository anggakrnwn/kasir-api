package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// create migrations table
func (m *Migrator) CreateTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(50) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := m.db.Exec(query)
	return err
}

// get applied migrations
func (m *Migrator) GetApplied() (map[string]bool, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, nil
}

// load migration files
func (m *Migrator) LoadFiles(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

// run migration
func (m *Migrator) Up() error {
	// create table dulu
	if err := m.CreateTable(); err != nil {
		return err
	}

	// load files
	files, err := m.LoadFiles("migrations/files")
	if err != nil {
		return err
	}

	// get applied migrations
	applied, err := m.GetApplied()
	if err != nil {
		return err
	}

	// run each file
	for _, file := range files {
		version := strings.TrimPrefix(filepath.Base(file), ".sql")
		version = strings.Split(version, "_")[0]

		if applied[version] {
			log.Printf("⏭️  Migration %s already applied, skipping", version)
			continue
		}

		// read SQL file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", file, err)
		}

		queries := strings.Split(string(content), ";")

		// begin transaction
		tx, err := m.db.Begin()
		if err != nil {
			return err
		}

		// execute each query
		log.Printf("applying migration: %s", filepath.Base(file))
		for _, query := range queries {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}

			if _, err := tx.Exec(query); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute %s: %v\nQuery: %s",
					filepath.Base(file), err, query)
			}
		}

		// record migration
		_, err = tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			version,
			filepath.Base(file),
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		// commit
		if err := tx.Commit(); err != nil {
			return err
		}

		log.Printf("migration %s applied", version)
	}

	return nil
}

// reset database (untuk development!)
func (m *Migrator) Reset() error {
	log.Println("RESET DATABASE - Menghapus semua tabel!")

	tables := []string{
		"transaction_details",
		"transactions",
		"products",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err := m.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Printf("error dropping %s: %v", table, err)
		}
	}

	log.Println("database reset, ready for fresh migration")
	return nil
}
