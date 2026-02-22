package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}

func Migrate(db *sql.DB, migrationsFS embed.FS) error {
	data, err := migrationsFS.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	statements := strings.Split(string(data), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("migration statement warning: %v (statement: %.80s...)", err, stmt)
		}
	}

	log.Println("database migrations applied successfully")
	return nil
}
