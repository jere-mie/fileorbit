package models

import (
	"database/sql"
	"time"
)

// File represents an uploaded file stored in the database.
type File struct {
	ID            int64          `db:"id" json:"id"`
	Filename      string         `db:"filename" json:"filename"`
	OriginalName  string         `db:"original_name" json:"original_name"`
	CustomURL     string         `db:"custom_url" json:"custom_url"`
	ContentType   string         `db:"content_type" json:"content_type"`
	Size          int64          `db:"size" json:"size"`
	Data          []byte         `db:"data" json:"-"`
	PasswordHash  sql.NullString `db:"password_hash" json:"-"`
	ExpiresAt     sql.NullTime   `db:"expires_at" json:"expires_at"`
	DownloadCount int64          `db:"download_count" json:"download_count"`
	Description   string         `db:"description" json:"description"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// AnalyticsEvent represents a single file access event.
type AnalyticsEvent struct {
	ID         int64     `db:"id" json:"id"`
	FileID     int64     `db:"file_id" json:"file_id"`
	AccessedAt time.Time `db:"accessed_at" json:"accessed_at"`
	Referrer   string    `db:"referrer" json:"referrer"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
}
