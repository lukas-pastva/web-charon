package handlers

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lukas-pastva/web-charon/internal/models"
)

type PublicHandler struct {
	Articles  *models.ArticleStore
	Galleries *models.GalleryStore
	Comments  *models.CommentStore
	Templates map[string]*template.Template
	BaseURL   string
	Version   string
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
		"BaseURL":         h.BaseURL,
		"CanonicalPath":   "/",
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

	canonicalPath := "/articles"
	if page > 1 {
		canonicalPath = "/articles?page=" + strconv.Itoa(page)
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"Page":          page,
		"TotalPages":    totalPages,
		"HasPrev":       page > 1,
		"HasNext":       page < totalPages,
		"PrevPage":      page - 1,
		"NextPage":      page + 1,
		"BaseURL":       h.BaseURL,
		"CanonicalPath": canonicalPath,
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

	data := map[string]interface{}{
		"Article":         article,
		"Comments":        comments,
		"Gallery":         gallery,
		"CommentsEnabled": article.CommentsEnabled,
		"BaseURL":         h.BaseURL,
		"CanonicalPath":   "/articles/" + article.Slug,
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
		"Galleries":     galleries,
		"BaseURL":       h.BaseURL,
		"CanonicalPath": "/gallery",
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
		"Gallery":       gallery,
		"BaseURL":       h.BaseURL,
		"CanonicalPath": "/gallery/" + gallery.Slug,
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

	if !article.CommentsEnabled {
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

func (h *PublicHandler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /admin/\n\nSitemap: %s/sitemap.xml\n", h.BaseURL)
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func (h *PublicHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	urls := []sitemapURL{
		{Loc: h.BaseURL + "/", ChangeFreq: "daily", Priority: "1.0"},
		{Loc: h.BaseURL + "/articles", ChangeFreq: "daily", Priority: "0.8"},
		{Loc: h.BaseURL + "/gallery", ChangeFreq: "weekly", Priority: "0.7"},
	}

	articles, err := h.Articles.GetPublished()
	if err == nil {
		for _, a := range articles {
			urls = append(urls, sitemapURL{
				Loc:        h.BaseURL + "/articles/" + a.Slug,
				LastMod:    a.UpdatedAt.Format(time.DateOnly),
				ChangeFreq: "monthly",
				Priority:   "0.6",
			})
		}
	}

	galleries, err := h.Galleries.GetAll()
	if err == nil {
		for _, g := range galleries {
			urls = append(urls, sitemapURL{
				Loc:        h.BaseURL + "/gallery/" + g.Slug,
				LastMod:    g.UpdatedAt.Format(time.DateOnly),
				ChangeFreq: "monthly",
				Priority:   "0.5",
			})
		}
	}

	sitemap := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	enc.Encode(sitemap)
}

func (h *PublicHandler) render(w http.ResponseWriter, name string, data interface{}) {
	t, ok := h.Templates[name]
	if !ok {
		log.Printf("public template not found: %s", name)
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	if m, ok := data.(map[string]interface{}); ok {
		m["Version"] = h.Version
	}
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("template error (%s): %v", name, err)
		http.Error(w, "Interní chyba serveru", 500)
	}
}
