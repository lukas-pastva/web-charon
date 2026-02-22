package router

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lukas-pastva/web-charon/internal/handlers"
)

func New(pub *handlers.PublicHandler, admin *handlers.AdminHandler, adminDomain, storagePath string, staticFS http.FileSystem) http.Handler {
	publicRouter := newPublicRouter(pub, storagePath, staticFS)
	adminRouter := newAdminRouter(admin)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		// Strip port if present
		if i := strings.LastIndex(host, ":"); i != -1 {
			host = host[:i]
		}

		adminHost := adminDomain
		if i := strings.LastIndex(adminHost, ":"); i != -1 {
			adminHost = adminHost[:i]
		}

		if host == adminHost {
			adminRouter.ServeHTTP(w, r)
		} else {
			publicRouter.ServeHTTP(w, r)
		}
	})
}

func newPublicRouter(h *handlers.PublicHandler, storagePath string, staticFS http.FileSystem) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", h.Home)
	r.Get("/articles", h.Articles_List)
	r.Get("/articles/{slug}", h.Article_Show)
	r.Post("/articles/{slug}/comments", h.Comment_Submit)
	r.Get("/gallery", h.Gallery_List)
	r.Get("/gallery/{slug}", h.Gallery_Show)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticFS)))

	// Uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(storagePath))))

	return r
}

func newAdminRouter(h *handlers.AdminHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", h.Dashboard)

	r.Get("/articles", h.Articles_List)
	r.Get("/articles/new", h.Articles_New)
	r.Post("/articles", h.Articles_Create)
	r.Get("/articles/{id}/edit", h.Articles_Edit)
	r.Post("/articles/{id}", h.Articles_Update)
	r.Post("/articles/{id}/delete", h.Articles_Delete)

	r.Get("/galleries", h.Galleries_List)
	r.Get("/galleries/new", h.Galleries_New)
	r.Post("/galleries", h.Galleries_Create)
	r.Get("/galleries/{id}/edit", h.Galleries_Edit)
	r.Post("/galleries/{id}", h.Galleries_Update)
	r.Post("/galleries/{id}/delete", h.Galleries_Delete)
	r.Post("/galleries/{id}/images", h.Galleries_UploadImages)

	r.Post("/images/{id}/delete", h.Images_Delete)

	r.Get("/comments", h.Comments_List)
	r.Post("/comments/{id}/approve", h.Comments_Approve)
	r.Post("/comments/{id}/delete", h.Comments_Delete)

	r.Get("/settings", h.Settings_Show)
	r.Post("/settings", h.Settings_Update)

	return r
}
