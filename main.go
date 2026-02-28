package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/jere-mie/fileorbit/internal/config"
	"github.com/jere-mie/fileorbit/internal/database"
	"github.com/jere-mie/fileorbit/internal/handlers"
	"github.com/jere-mie/fileorbit/internal/middleware"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

//go:embed version.txt
var version string

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Handle CLI commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Print(version)
			return
		case "migrate":
			if len(os.Args) > 2 && os.Args[2] == "status" {
				if err := database.MigrationStatus(db); err != nil {
					log.Fatal(err)
				}
				return
			}
			if err := database.RunMigrations(db); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Migrations completed successfully")
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: fileorbit [version | migrate [status]]\n")
			os.Exit(1)
		}
	}

	// Auto-run migrations on startup
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	auth := middleware.NewAuthMiddleware(cfg.SessionSecret)
	app := handlers.NewApp(db, cfg, auth, templateFS)

	mux := http.NewServeMux()

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Auth routes
	mux.HandleFunc("GET /login", app.LoginPage)
	mux.HandleFunc("POST /login", app.LoginHandler)
	mux.HandleFunc("GET /logout", app.LogoutHandler)

	// Root redirect
	mux.HandleFunc("GET /{$}", app.RootHandler)

	// Protected dashboard routes
	mux.Handle("GET /dashboard", auth.Require(http.HandlerFunc(app.Dashboard)))
	mux.Handle("POST /dashboard/upload", auth.Require(http.HandlerFunc(app.UploadFile)))
	mux.Handle("GET /dashboard/files/{id}/edit", auth.Require(http.HandlerFunc(app.EditFilePage)))
	mux.Handle("POST /dashboard/files/{id}/edit", auth.Require(http.HandlerFunc(app.EditFileHandler)))
	mux.Handle("POST /dashboard/files/{id}/delete", auth.Require(http.HandlerFunc(app.DeleteFile)))
	mux.Handle("GET /dashboard/files/{id}/analytics", auth.Require(http.HandlerFunc(app.FileAnalytics)))
	mux.Handle("GET /api/search", auth.Require(http.HandlerFunc(app.SearchFiles)))

	// Public file access (catch-all, must be registered last)
	mux.HandleFunc("GET /{path...}", app.ServeFile)
	mux.HandleFunc("POST /{path...}", app.ServeFileWithPassword)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("FileOrbit starting on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
