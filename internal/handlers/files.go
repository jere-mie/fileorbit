package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jere-mie/fileorbit/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// DashboardData holds template data for the dashboard page.
type DashboardData struct {
	Files       []models.File
	Search      string
	Success     string
	Error       string
	BaseURL     string
	MaxFileSize int64
}

// EditData holds template data for the edit file page.
type EditData struct {
	File    models.File
	Error   string
	Success string
	BaseURL string
}

const fileListQuery = `SELECT id, filename, original_name, custom_url, content_type, size,
	password_hash, expires_at, download_count, description, created_at, updated_at
	FROM files ORDER BY created_at DESC`

const fileFindQuery = `SELECT id, filename, original_name, custom_url, content_type, size,
	password_hash, expires_at, download_count, description, created_at, updated_at
	FROM files WHERE id = ?`

// Dashboard renders the main dashboard with file listing.
func (a *App) Dashboard(w http.ResponseWriter, r *http.Request) {
	var files []models.File
	if err := a.DB.Select(&files, fileListQuery); err != nil {
		http.Error(w, "Failed to load files", http.StatusInternalServerError)
		return
	}

	success := ""
	switch r.URL.Query().Get("success") {
	case "uploaded":
		success = "File uploaded successfully"
	case "updated":
		success = "File updated successfully"
	case "deleted":
		success = "File deleted successfully"
	}

	a.render(w, "dashboard", DashboardData{
		Files:       files,
		Success:     success,
		BaseURL:     a.Config.BaseURL,
		MaxFileSize: a.Config.MaxFileSize,
	})
}

// UploadFile handles file upload form submissions.
func (a *App) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.Config.MaxFileSize+1024*1024)
	if err := r.ParseMultipartForm(a.Config.MaxFileSize); err != nil {
		a.dashboardWithError(w, fmt.Sprintf("File too large. Maximum size is %s", formatBytes(a.Config.MaxFileSize)))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		a.dashboardWithError(w, "No file selected")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		a.dashboardWithError(w, "Failed to read file")
		return
	}

	customURL := strings.TrimSpace(r.FormValue("custom_url"))
	description := strings.TrimSpace(r.FormValue("description"))
	password := strings.TrimSpace(r.FormValue("password"))
	expiresStr := strings.TrimSpace(r.FormValue("expires_at"))

	if customURL == "" {
		customURL = generateCustomURL(header.Filename)
	}

	customURL = strings.TrimPrefix(customURL, "/")

	// Check for reserved paths
	reserved := []string{"login", "logout", "dashboard", "api", "static"}
	for _, res := range reserved {
		if strings.HasPrefix(strings.ToLower(customURL), res) {
			a.dashboardWithError(w, "Custom URL cannot start with reserved path: "+res)
			return
		}
	}

	// Check uniqueness
	var count int
	if err := a.DB.Get(&count, "SELECT COUNT(*) FROM files WHERE custom_url = ?", customURL); err != nil {
		a.dashboardWithError(w, "Database error")
		return
	}
	if count > 0 {
		a.dashboardWithError(w, "Custom URL already in use")
		return
	}

	// Hash password if provided
	var passwordHash sql.NullString
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			a.dashboardWithError(w, "Failed to process password")
			return
		}
		passwordHash = sql.NullString{String: string(hash), Valid: true}
	}

	// Parse expiration
	var expiresAt sql.NullTime
	if expiresStr != "" {
		t, err := time.Parse("2006-01-02T15:04", expiresStr)
		if err == nil {
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectContentType(header.Filename, data)
	}

	_, err = a.DB.Exec(
		`INSERT INTO files (filename, original_name, custom_url, content_type, size, data, password_hash, expires_at, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		header.Filename, header.Filename, customURL, contentType, len(data), data, passwordHash, expiresAt, description,
	)
	if err != nil {
		a.dashboardWithError(w, "Failed to save file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/dashboard?success=uploaded", http.StatusSeeOther)
}

// EditFilePage renders the file editing form.
func (a *App) EditFilePage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var file models.File
	if err := a.DB.Get(&file, fileFindQuery, id); err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	a.render(w, "edit", EditData{
		File:    file,
		BaseURL: a.Config.BaseURL,
	})
}

// EditFileHandler processes file metadata edit submissions.
func (a *App) EditFileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var file models.File
	if err := a.DB.Get(&file, fileFindQuery, id); err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	customURL := strings.TrimSpace(r.FormValue("custom_url"))
	description := strings.TrimSpace(r.FormValue("description"))
	filename := strings.TrimSpace(r.FormValue("filename"))
	password := strings.TrimSpace(r.FormValue("password"))
	removePassword := r.FormValue("remove_password") == "on"
	expiresStr := strings.TrimSpace(r.FormValue("expires_at"))
	removeExpiry := r.FormValue("remove_expiry") == "on"

	if customURL == "" {
		customURL = file.CustomURL
	}
	if filename == "" {
		filename = file.Filename
	}

	customURL = strings.TrimPrefix(customURL, "/")

	// Check uniqueness if URL changed
	if customURL != file.CustomURL {
		var count int
		if err := a.DB.Get(&count, "SELECT COUNT(*) FROM files WHERE custom_url = ? AND id != ?", customURL, file.ID); err != nil {
			a.render(w, "edit", EditData{File: file, Error: "Database error", BaseURL: a.Config.BaseURL})
			return
		}
		if count > 0 {
			a.render(w, "edit", EditData{File: file, Error: "Custom URL already in use", BaseURL: a.Config.BaseURL})
			return
		}
	}

	// Handle password
	var passwordHash sql.NullString
	if removePassword {
		passwordHash = sql.NullString{Valid: false}
	} else if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			a.render(w, "edit", EditData{File: file, Error: "Failed to process password", BaseURL: a.Config.BaseURL})
			return
		}
		passwordHash = sql.NullString{String: string(hash), Valid: true}
	} else {
		passwordHash = file.PasswordHash
	}

	// Handle expiration
	var expiresAt sql.NullTime
	if removeExpiry {
		expiresAt = sql.NullTime{Valid: false}
	} else if expiresStr != "" {
		t, err := time.Parse("2006-01-02T15:04", expiresStr)
		if err == nil {
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
	} else {
		expiresAt = file.ExpiresAt
	}

	_, err := a.DB.Exec(
		`UPDATE files SET filename = ?, custom_url = ?, description = ?, password_hash = ?, expires_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		filename, customURL, description, passwordHash, expiresAt, file.ID,
	)
	if err != nil {
		a.render(w, "edit", EditData{File: file, Error: "Failed to update file", BaseURL: a.Config.BaseURL})
		return
	}

	http.Redirect(w, r, "/dashboard?success=updated", http.StatusSeeOther)
}

// DeleteFile removes a file from the database.
func (a *App) DeleteFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := a.DB.Exec("DELETE FROM files WHERE id = ?", id); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard?success=deleted", http.StatusSeeOther)
}

// SearchFiles returns a JSON list of files matching a query.
func (a *App) SearchFiles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	w.Header().Set("Content-Type", "application/json")

	if query == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"files": []interface{}{}})
		return
	}

	searchTerm := "%" + query + "%"
	var files []models.File
	err := a.DB.Select(&files,
		`SELECT id, filename, original_name, custom_url, content_type, size,
			password_hash, expires_at, download_count, description, created_at, updated_at
		FROM files
		WHERE filename LIKE ? OR description LIKE ? OR custom_url LIKE ?
		ORDER BY created_at DESC`,
		searchTerm, searchTerm, searchTerm,
	)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Search failed"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"files": files})
}

func (a *App) dashboardWithError(w http.ResponseWriter, errMsg string) {
	var files []models.File
	a.DB.Select(&files, fileListQuery)
	a.render(w, "dashboard", DashboardData{
		Files:       files,
		Error:       errMsg,
		BaseURL:     a.Config.BaseURL,
		MaxFileSize: a.Config.MaxFileSize,
	})
}

func generateCustomURL(filename string) string {
	ext := filepath.Ext(filename)
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes) + ext
}

func detectContentType(filename string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".pdf":  "application/pdf",
		".html": "text/html",
		".htm":  "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".md":   "text/markdown",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".rar":  "application/vnd.rar",
		".7z":   "application/x-7z-compressed",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".wasm": "application/wasm",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	return http.DetectContentType(data)
}
