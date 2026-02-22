package main

import (
	"crypto/rand"
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
	"golang.org/x/crypto/bcrypt"
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

	// Generate session secret
	sessionSecret := make([]byte, 32)
	if _, err := rand.Read(sessionSecret); err != nil {
		log.Fatalf("failed to generate session secret: %v", err)
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
	userStore := &models.UserStore{DB: db}

	// Seed initial admin user if no users exist
	count, err := userStore.Count()
	if err != nil {
		log.Printf("warning: could not count users: %v", err)
	} else if count == 0 {
		adminPass := cfg.AdminPassword
		if adminPass == "" {
			adminPass = "admin"
			log.Println("WARNING: No ADMIN_PASSWORD set, using default password 'admin'")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash admin password: %v", err)
		}
		adminUser := &models.User{
			Name:         "Admin",
			Surname:      "",
			Nickname:     "admin",
			PasswordHash: string(hash),
			IsAdmin:      true,
		}
		if err := userStore.Create(adminUser); err != nil {
			log.Fatalf("failed to create initial admin user: %v", err)
		}
		log.Println("created initial admin user (nickname: admin)")
	}

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
		Users:       userStore,
		Templates:   adminTmpl,
		StoragePath: cfg.StoragePath,
	}

	authHandler := &handlers.AuthHandler{
		Users:         userStore,
		Templates:     adminTmpl,
		SessionSecret: sessionSecret,
	}

	// Static file system
	staticSub, err := fs.Sub(charon.StaticFS, "static")
	if err != nil {
		log.Fatalf("failed to create static sub filesystem: %v", err)
	}

	// Create router
	handler := router.New(publicHandler, adminHandler, authHandler, cfg.StoragePath, http.FS(staticSub))

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Charon starting on %s (public: %s)", addr, cfg.PublicDomain)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
