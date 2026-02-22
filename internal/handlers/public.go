package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lukas-pastva/web-charon/internal/models"
)

type PublicHandler struct {
	Articles  *models.ArticleStore
	Galleries *models.GalleryStore
	Comments  *models.CommentStore
	Settings  *models.SettingsStore
	Templates *template.Template
}

func (h *PublicHandler) Home(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.Articles.GetPublished()
	limit := 6
	if len(articles) > limit {
		articles = articles[:limit]
	}

	galleries, _ := h.Galleries.GetAll()
	var featuredGallery *models.Gallery
	if len(galleries) > 0 {
		g, err := h.Galleries.GetByID(galleries[0].ID)
		if err == nil {
			featuredGallery = g
		}
	}

	data := map[string]interface{}{
		"Articles":        articles,
		"FeaturedGallery": featuredGallery,
	}
	h.render(w, "home.html", data)
}

func (h *PublicHandler) Articles_List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage := 10
	offset := (page - 1) * perPage

	articles, total, err := h.Articles.GetPublishedPaginated(perPage, offset)
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}

	totalPages := (total + perPage - 1) / perPage

	data := map[string]interface{}{
		"Articles":   articles,
		"Page":       page,
		"TotalPages": totalPages,
		"HasPrev":    page > 1,
		"HasNext":    page < totalPages,
		"PrevPage":   page - 1,
		"NextPage":   page + 1,
	}
	h.render(w, "articles.html", data)
}

func (h *PublicHandler) Article_Show(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	article, err := h.Articles.GetBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !article.Published {
		http.NotFound(w, r)
		return
	}

	comments, _ := h.Comments.GetByArticleID(article.ID, true)

	gallery, err := h.Galleries.GetByArticleID(article.ID)
	if err == sql.ErrNoRows {
		gallery = nil
	}

	commentsEnabled, _ := h.Settings.Get("comments_enabled")

	data := map[string]interface{}{
		"Article":         article,
		"Comments":        comments,
		"Gallery":         gallery,
		"CommentsEnabled": commentsEnabled == "true",
	}
	h.render(w, "article.html", data)
}

func (h *PublicHandler) Gallery_List(w http.ResponseWriter, r *http.Request) {
	galleries, err := h.Galleries.GetAll()
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}

	// Load first image for each gallery as cover
	for i := range galleries {
		images, err := h.Galleries.GetImages(galleries[i].ID)
		if err == nil {
			galleries[i].Images = images
		}
	}

	data := map[string]interface{}{
		"Galleries": galleries,
	}
	h.render(w, "gallery.html", data)
}

func (h *PublicHandler) Gallery_Show(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	gallery, err := h.Galleries.GetBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Gallery": gallery,
	}
	h.render(w, "gallery_detail.html", data)
}

func (h *PublicHandler) Comment_Submit(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	article, err := h.Articles.GetBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	commentsEnabled, _ := h.Settings.Get("comments_enabled")
	if commentsEnabled != "true" {
		http.Error(w, "Komentáře jsou zakázány", http.StatusForbidden)
		return
	}

	authorName := strings.TrimSpace(r.FormValue("author_name"))
	content := strings.TrimSpace(r.FormValue("content"))

	if authorName == "" || content == "" {
		http.Redirect(w, r, "/articles/"+slug+"?error=fields_required", http.StatusSeeOther)
		return
	}

	comment := &models.Comment{
		ArticleID:  article.ID,
		AuthorName: authorName,
		Content:    content,
		Approved:   false,
	}
	if err := h.Comments.Create(comment); err != nil {
		log.Printf("error creating comment: %v", err)
		http.Error(w, "Interní chyba serveru", 500)
		return
	}

	http.Redirect(w, r, "/articles/"+slug+"?comment=pending", http.StatusSeeOther)
}

func (h *PublicHandler) render(w http.ResponseWriter, name string, data interface{}) {
	err := h.Templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("template error (%s): %v", name, err)
		http.Error(w, "Interní chyba serveru", 500)
	}
}
