package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Migration represents a single database migration.
type Migration struct {
	Version int
	Name    string
	Up      string
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_files_table",
		Up: `CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			custom_url TEXT NOT NULL UNIQUE,
			content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			size INTEGER NOT NULL DEFAULT 0,
			data BLOB,
			password_hash TEXT,
			expires_at DATETIME,
			download_count INTEGER NOT NULL DEFAULT 0,
			description TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	},
	{
		Version: 2,
		Name:    "create_analytics_table",
		Up: `CREATE TABLE IF NOT EXISTS analytics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			accessed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			referrer TEXT NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			ip_address TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
		)`,
	},
	{
		Version: 3,
		Name:    "create_indexes",
		Up: `CREATE INDEX IF NOT EXISTS idx_files_custom_url ON files(custom_url);
			 CREATE INDEX IF NOT EXISTS idx_analytics_file_id ON analytics(file_id);
			 CREATE INDEX IF NOT EXISTS idx_analytics_accessed_at ON analytics(accessed_at)`,
	},
}

// RunMigrations executes all pending database migrations.
func RunMigrations(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	applied := make(map[int]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("querying migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return err
		}
		applied[version] = true
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		log.Printf("Running migration %d: %s", m.Version, m.Name)

		if _, err := db.Exec(m.Up); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		if _, err := db.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			m.Version, m.Name,
		); err != nil {
			return fmt.Errorf("recording migration %d: %w", m.Version, err)
		}

		log.Printf("Migration %d applied successfully", m.Version)
	}

	return nil
}

// MigrationStatus prints the current state of all migrations.
func MigrationStatus(db *sqlx.DB) error {
	type status struct {
		Version   int       `db:"version"`
		Name      string    `db:"name"`
		AppliedAt time.Time `db:"applied_at"`
	}

	var statuses []status
	err := db.Select(&statuses, "SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		fmt.Println("No migrations have been applied yet.")
		return nil
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations have been applied yet.")
		return nil
	}

	fmt.Printf("%-10s %-30s %-30s\n", "VERSION", "NAME", "APPLIED AT")
	fmt.Println(strings.Repeat("-", 70))
	for _, s := range statuses {
		fmt.Printf("%-10d %-30s %-30s\n", s.Version, s.Name, s.AppliedAt.Format(time.RFC3339))
	}

	pending := 0
	for _, m := range migrations {
		found := false
		for _, s := range statuses {
			if s.Version == m.Version {
				found = true
				break
			}
		}
		if !found {
			pending++
		}
	}

	fmt.Printf("\n%d applied, %d pending\n", len(statuses), pending)
	return nil
}
