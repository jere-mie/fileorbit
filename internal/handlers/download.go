package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jere-mie/fileorbit/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// PasswordPromptData holds template data for the password prompt page.
type PasswordPromptData struct {
	CustomURL string
	Error     string
	Filename  string
}

// ServeFile handles public file downloads via custom URLs.
func (a *App) ServeFile(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if path == "" {
		a.RootHandler(w, r)
		return
	}

	// Query without loading blob data
	var file models.File
	err := a.DB.Get(&file,
		`SELECT id, filename, original_name, custom_url, content_type, size,
			password_hash, expires_at, download_count, description, created_at, updated_at
		FROM files WHERE custom_url = ?`, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check expiration
	if file.ExpiresAt.Valid && file.ExpiresAt.Time.Before(time.Now()) {
		http.Error(w, "This file has expired", http.StatusGone)
		return
	}

	// Check if password protected
	if file.PasswordHash.Valid && file.PasswordHash.String != "" {
		a.render(w, "password_prompt", PasswordPromptData{
			CustomURL: file.CustomURL,
			Filename:  file.Filename,
		})
		return
	}

	// Load blob data and serve
	var data []byte
	if err := a.DB.Get(&data, "SELECT data FROM files WHERE id = ?", file.ID); err != nil {
		http.Error(w, "Failed to load file data", http.StatusInternalServerError)
		return
	}
	file.Data = data

	a.serveFileContent(w, r, &file)
}

// ServeFileWithPassword handles password-protected file downloads.
func (a *App) ServeFileWithPassword(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	var file models.File
	err := a.DB.Get(&file,
		`SELECT id, filename, original_name, custom_url, content_type, size,
			password_hash, expires_at, download_count, description, created_at, updated_at
		FROM files WHERE custom_url = ?`, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Check expiration
	if file.ExpiresAt.Valid && file.ExpiresAt.Time.Before(time.Now()) {
		http.Error(w, "This file has expired", http.StatusGone)
		return
	}

	password := r.FormValue("password")
	if err := bcrypt.CompareHashAndPassword([]byte(file.PasswordHash.String), []byte(password)); err != nil {
		a.render(w, "password_prompt", PasswordPromptData{
			CustomURL: file.CustomURL,
			Filename:  file.Filename,
			Error:     "Incorrect password",
		})
		return
	}

	// Load blob data and serve
	var data []byte
	if err := a.DB.Get(&data, "SELECT data FROM files WHERE id = ?", file.ID); err != nil {
		http.Error(w, "Failed to load file data", http.StatusInternalServerError)
		return
	}
	file.Data = data

	a.serveFileContent(w, r, &file)
}

func (a *App) serveFileContent(w http.ResponseWriter, r *http.Request, file *models.File) {
	// Record analytics
	referrer := r.Header.Get("Referer")
	userAgent := r.Header.Get("User-Agent")
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}

	a.DB.Exec(
		`INSERT INTO analytics (file_id, referrer, user_agent, ip_address) VALUES (?, ?, ?, ?)`,
		file.ID, referrer, userAgent, ip,
	)

	a.DB.Exec("UPDATE files SET download_count = download_count + 1 WHERE id = ?", file.ID)

	// Set response headers
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	if isViewableType(file.ContentType) {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Filename))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Filename))
	}

	w.Write(file.Data)
}

func isViewableType(contentType string) bool {
	viewable := []string{"text/", "image/", "application/pdf", "video/", "audio/"}
	for _, v := range viewable {
		if strings.HasPrefix(contentType, v) {
			return true
		}
	}
	return false
}
