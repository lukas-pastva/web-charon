package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lukas-pastva/web-charon/internal/handlers"
)

func New(pub *handlers.PublicHandler, admin *handlers.AdminHandler, auth *handlers.AuthHandler, storagePath string, staticFS http.FileSystem) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/", pub.Home)
	r.Get("/articles", pub.Articles_List)
	r.Get("/articles/{slug}", pub.Article_Show)
	r.Post("/articles/{slug}/comments", pub.Comment_Submit)
	r.Get("/gallery", pub.Gallery_List)
	r.Get("/gallery/{slug}", pub.Gallery_Show)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticFS)))

	// Uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(storagePath))))

	// Admin routes
	r.Route("/admin", func(r chi.Router) {
		// Unauthenticated routes
		r.Get("/login", auth.Login)
		r.Post("/login", auth.LoginPost)
		r.Post("/logout", auth.Logout)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth)

			r.Get("/", admin.Dashboard)

			// Profile (any authenticated user)
			r.Get("/profile", admin.Profile_Show)
			r.Post("/profile", admin.Profile_Update)

			r.Get("/articles", admin.Articles_List)
			r.Get("/articles/new", admin.Articles_New)
			r.Post("/articles", admin.Articles_Create)
			r.Get("/articles/{id}/edit", admin.Articles_Edit)
			r.Post("/articles/{id}", admin.Articles_Update)
			r.Post("/articles/{id}/delete", admin.Articles_Delete)

			r.Get("/galleries", admin.Galleries_List)
			r.Get("/galleries/new", admin.Galleries_New)
			r.Post("/galleries", admin.Galleries_Create)
			r.Get("/galleries/{id}/edit", admin.Galleries_Edit)
			r.Post("/galleries/{id}", admin.Galleries_Update)
			r.Post("/galleries/{id}/delete", admin.Galleries_Delete)
			r.Post("/galleries/{id}/images", admin.Galleries_UploadImages)

			r.Post("/images/{id}/delete", admin.Images_Delete)

			r.Get("/comments", admin.Comments_List)
			r.Post("/comments/{id}/approve", admin.Comments_Approve)
			r.Post("/comments/{id}/delete", admin.Comments_Delete)

			r.Get("/settings", admin.Settings_Show)
			r.Post("/settings", admin.Settings_Update)

			// User management (admin-only)
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAdmin)

				r.Get("/users", admin.Users_List)
				r.Get("/users/new", admin.Users_New)
				r.Post("/users", admin.Users_Create)
				r.Get("/users/{id}/edit", admin.Users_Edit)
				r.Post("/users/{id}", admin.Users_Update)
				r.Post("/users/{id}/delete", admin.Users_Delete)
			})
		})
	})

	return r
}
