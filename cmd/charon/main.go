package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	charon "github.com/lukas-pastva/web-charon"
	"github.com/lukas-pastva/web-charon/internal/config"
	"github.com/lukas-pastva/web-charon/internal/database"
	"github.com/lukas-pastva/web-charon/internal/handlers"
	"github.com/lukas-pastva/web-charon/internal/models"
	"github.com/lukas-pastva/web-charon/internal/router"
)

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db, charon.MigrationsFS); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}

	// Parse templates
	funcMap := template.FuncMap{
		"nl2br": func(s string) template.HTML {
			return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
		},
		"deref": func(p *int64) int64 {
			if p == nil {
				return 0
			}
			return *p
		},
	}

	publicTmpl, err := template.New("").Funcs(funcMap).ParseFS(charon.PublicTemplatesFS, "templates/public/*.html")
	if err != nil {
		log.Fatalf("failed to parse public templates: %v", err)
	}

	adminTmpl, err := template.New("").Funcs(funcMap).ParseFS(charon.AdminTemplatesFS, "templates/admin/*.html")
	if err != nil {
		log.Fatalf("failed to parse admin templates: %v", err)
	}

	// Initialize stores
	articleStore := &models.ArticleStore{DB: db}
	galleryStore := &models.GalleryStore{DB: db}
	commentStore := &models.CommentStore{DB: db}
	settingsStore := &models.SettingsStore{DB: db}

	// Initialize handlers
	publicHandler := &handlers.PublicHandler{
		Articles:  articleStore,
		Galleries: galleryStore,
		Comments:  commentStore,
		Settings:  settingsStore,
		Templates: publicTmpl,
	}

	adminHandler := &handlers.AdminHandler{
		Articles:    articleStore,
		Galleries:   galleryStore,
		Comments:    commentStore,
		Settings:    settingsStore,
		Templates:   adminTmpl,
		StoragePath: cfg.StoragePath,
	}

	// Static file system
	staticSub, err := fs.Sub(charon.StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub filesystem: %v", err)
	}

	// Create router
	handler := router.New(publicHandler, adminHandler, cfg.AdminDomain, cfg.StoragePath, http.FS(staticSub))

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Charon starting on %s (public: %s, admin: %s)", addr, cfg.PublicDomain, cfg.AdminDomain)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
