package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lukas-pastva/web-charon/internal/models"
)

type AdminHandler struct {
	Articles    *models.ArticleStore
	Galleries   *models.GalleryStore
	Comments    *models.CommentStore
	Settings    *models.SettingsStore
	Templates   *template.Template
	StoragePath string
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.Articles.GetAll()
	galleries, _ := h.Galleries.GetAll()
	pending, _ := h.Comments.GetAllPending()

	data := map[string]interface{}{
		"ArticleCount": len(articles),
		"GalleryCount": len(galleries),
		"PendingCount": len(pending),
	}
	h.render(w, "dashboard.html", data)
}

// --- Articles ---

func (h *AdminHandler) Articles_List(w http.ResponseWriter, r *http.Request) {
	articles, err := h.Articles.GetAll()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.render(w, "articles.html", map[string]interface{}{"Articles": articles})
}

func (h *AdminHandler) Articles_New(w http.ResponseWriter, r *http.Request) {
	h.render(w, "article_form.html", map[string]interface{}{"Article": &models.Article{}, "IsNew": true})
}

func (h *AdminHandler) Articles_Create(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	article := &models.Article{
		Title:   strings.TrimSpace(r.FormValue("title")),
		Slug:    strings.TrimSpace(r.FormValue("slug")),
		Content: r.FormValue("content"),
		Excerpt: strings.TrimSpace(r.FormValue("excerpt")),
	}
	article.Published = r.FormValue("published") == "on"

	if f, _, err := r.FormFile("cover_image"); err == nil {
		f.Close()
		filename, err := HandleUpload(r, "cover_image", h.StoragePath)
		if err == nil {
			article.CoverImage = filename
		}
	}

	if err := h.Articles.Create(article); err != nil {
		log.Printf("error creating article: %v", err)
		h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": true, "Error": "Failed to create article. Ensure slug is unique."})
		return
	}

	http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
}

func (h *AdminHandler) Articles_Edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	article, err := h.Articles.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": false})
}

func (h *AdminHandler) Articles_Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	article, err := h.Articles.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseMultipartForm(32 << 20)

	article.Title = strings.TrimSpace(r.FormValue("title"))
	article.Slug = strings.TrimSpace(r.FormValue("slug"))
	article.Content = r.FormValue("content")
	article.Excerpt = strings.TrimSpace(r.FormValue("excerpt"))
	article.Published = r.FormValue("published") == "on"

	if f, _, err := r.FormFile("cover_image"); err == nil {
		f.Close()
		filename, err := HandleUpload(r, "cover_image", h.StoragePath)
		if err == nil {
			article.CoverImage = filename
		}
	}

	if err := h.Articles.Update(article); err != nil {
		log.Printf("error updating article: %v", err)
		h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": false, "Error": "Failed to update article."})
		return
	}

	http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
}

func (h *AdminHandler) Articles_Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.Articles.Delete(id)
	http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
}

// --- Galleries ---

func (h *AdminHandler) Galleries_List(w http.ResponseWriter, r *http.Request) {
	galleries, err := h.Galleries.GetAll()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	for i := range galleries {
		images, _ := h.Galleries.GetImages(galleries[i].ID)
		galleries[i].Images = images
	}
	h.render(w, "galleries.html", map[string]interface{}{"Galleries": galleries})
}

func (h *AdminHandler) Galleries_New(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.Articles.GetAll()
	h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": &models.Gallery{}, "IsNew": true, "Articles": articles})
}

func (h *AdminHandler) Galleries_Create(w http.ResponseWriter, r *http.Request) {
	gallery := &models.Gallery{
		Title:       strings.TrimSpace(r.FormValue("title")),
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Description: r.FormValue("description"),
	}

	if aid := r.FormValue("article_id"); aid != "" {
		id, err := strconv.ParseInt(aid, 10, 64)
		if err == nil && id > 0 {
			gallery.ArticleID = &id
		}
	}

	if err := h.Galleries.Create(gallery); err != nil {
		log.Printf("error creating gallery: %v", err)
		articles, _ := h.Articles.GetAll()
		h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": gallery, "IsNew": true, "Articles": articles, "Error": "Failed to create gallery. Ensure slug is unique."})
		return
	}

	http.Redirect(w, r, "/admin/galleries", http.StatusSeeOther)
}

func (h *AdminHandler) Galleries_Edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	gallery, err := h.Galleries.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	articles, _ := h.Articles.GetAll()
	h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": gallery, "IsNew": false, "Articles": articles})
}

func (h *AdminHandler) Galleries_Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	gallery, err := h.Galleries.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	gallery.Title = strings.TrimSpace(r.FormValue("title"))
	gallery.Slug = strings.TrimSpace(r.FormValue("slug"))
	gallery.Description = r.FormValue("description")
	gallery.ArticleID = nil

	if aid := r.FormValue("article_id"); aid != "" {
		aidInt, err := strconv.ParseInt(aid, 10, 64)
		if err == nil && aidInt > 0 {
			gallery.ArticleID = &aidInt
		}
	}

	if err := h.Galleries.Update(gallery); err != nil {
		log.Printf("error updating gallery: %v", err)
	}

	http.Redirect(w, r, "/admin/galleries", http.StatusSeeOther)
}

func (h *AdminHandler) Galleries_Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.Galleries.Delete(id)
	http.Redirect(w, r, "/admin/galleries", http.StatusSeeOther)
}

func (h *AdminHandler) Galleries_UploadImages(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	_, err := h.Galleries.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseMultipartForm(64 << 20)
	files := r.MultipartForm.File["images"]

	for i, fh := range files {
		file, err := fh.Open()
		if err != nil {
			continue
		}
		file.Close()

		filename, err := HandleUploadFromFileHeader(fh, h.StoragePath)
		if err != nil {
			log.Printf("upload error: %v", err)
			continue
		}

		img := &models.Image{
			GalleryID: id,
			Filename:  filename,
			Caption:   "",
			SortOrder: i,
		}
		h.Galleries.AddImage(img)
	}

	http.Redirect(w, r, "/admin/galleries/"+strconv.FormatInt(id, 10)+"/edit", http.StatusSeeOther)
}

func (h *AdminHandler) Images_Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	img, err := h.Galleries.GetImageByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	galleryID := img.GalleryID
	h.Galleries.DeleteImage(id)
	http.Redirect(w, r, "/admin/galleries/"+strconv.FormatInt(galleryID, 10)+"/edit", http.StatusSeeOther)
}

// --- Comments ---

func (h *AdminHandler) Comments_List(w http.ResponseWriter, r *http.Request) {
	comments, err := h.Comments.GetAll()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.render(w, "comments.html", map[string]interface{}{"Comments": comments})
}

func (h *AdminHandler) Comments_Approve(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.Comments.Approve(id)
	http.Redirect(w, r, "/admin/comments", http.StatusSeeOther)
}

func (h *AdminHandler) Comments_Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.Comments.Delete(id)
	http.Redirect(w, r, "/admin/comments", http.StatusSeeOther)
}

// --- Settings ---

func (h *AdminHandler) Settings_Show(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Settings.GetAll()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	saved := r.URL.Query().Get("saved") == "true"
	h.render(w, "settings.html", map[string]interface{}{"Settings": settings, "Saved": saved})
}

func (h *AdminHandler) Settings_Update(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	commentsEnabled := "false"
	if r.FormValue("comments_enabled") == "on" {
		commentsEnabled = "true"
	}
	h.Settings.Set("comments_enabled", commentsEnabled)
	http.Redirect(w, r, "/admin/settings?saved=true", http.StatusSeeOther)
}

func (h *AdminHandler) render(w http.ResponseWriter, name string, data interface{}) {
	err := h.Templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("admin template error (%s): %v", name, err)
		http.Error(w, "Internal Server Error", 500)
	}
}
