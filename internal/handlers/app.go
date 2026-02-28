package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jere-mie/fileorbit/internal/config"
	"github.com/jere-mie/fileorbit/internal/middleware"
	"github.com/jmoiron/sqlx"
)

// App holds application state and dependencies.
type App struct {
	DB        *sqlx.DB
	Config    *config.Config
	Auth      *middleware.AuthMiddleware
	templates map[string]*template.Template
}

// NewApp creates a new App instance with parsed templates.
func NewApp(db *sqlx.DB, cfg *config.Config, auth *middleware.AuthMiddleware, templateFS embed.FS) *App {
	app := &App{
		DB:        db,
		Config:    cfg,
		Auth:      auth,
		templates: make(map[string]*template.Template),
	}

	funcMap := template.FuncMap{
		"formatBytes": formatBytes,
		"formatDate":  formatDate,
		"formatDateTime": func(t time.Time) string {
			return t.Format("Jan 02, 2006 15:04")
		},
		"truncate": func(s string, length int) string {
			runes := []rune(s)
			if len(runes) <= length {
				return s
			}
			return string(runes[:length]) + "…"
		},
		"fileExt": func(name string) string {
			ext := filepath.Ext(name)
			if ext == "" {
				return "FILE"
			}
			return strings.ToUpper(ext[1:])
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"maxCount": func(daily []DailyCount) int {
			m := 0
			for _, d := range daily {
				if d.Count > m {
					m = d.Count
				}
			}
			if m == 0 {
				return 1
			}
			return m
		},
		"barHeight": func(count, max int) int {
			if max == 0 {
				return 0
			}
			h := (count * 100) / max
			if h < 4 {
				h = 4
			}
			return h
		},
		"shortDate": func(date string) string {
			t, err := time.Parse("2006-01-02", date)
			if err != nil {
				return date
			}
			return t.Format("Jan 2")
		},
	}

	pages := []string{"login", "dashboard", "edit", "file_analytics", "password_prompt"}
	for _, page := range pages {
		tmpl := template.Must(
			template.New("").Funcs(funcMap).ParseFS(
				templateFS,
				"templates/base.html",
				"templates/"+page+".html",
			),
		)
		app.templates[page] = tmpl
	}

	return app
}

// render executes a named template with the given data.
func (a *App) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := a.templates[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}
