package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// Connect opens a connection to the SQLite database and configures it.
func Connect(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Configure SQLite for better performance and safety
	db.MustExec("PRAGMA journal_mode=WAL")
	db.MustExec("PRAGMA foreign_keys=ON")
	db.MustExec("PRAGMA busy_timeout=5000")

	return db, nil
}
