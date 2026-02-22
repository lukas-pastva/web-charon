package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lukas-pastva/web-charon/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	Articles    *models.ArticleStore
	Galleries   *models.GalleryStore
	Comments    *models.CommentStore
	Settings    *models.SettingsStore
	Users       *models.UserStore
	Templates   map[string]*template.Template
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
		"CurrentUser":  CurrentUser(r),
	}
	h.render(w, "dashboard.html", data)
}

// --- Articles ---

func (h *AdminHandler) Articles_List(w http.ResponseWriter, r *http.Request) {
	articles, err := h.Articles.GetAll()
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	h.render(w, "articles.html", map[string]interface{}{"Articles": articles, "CurrentUser": CurrentUser(r)})
}

func (h *AdminHandler) Articles_New(w http.ResponseWriter, r *http.Request) {
	h.render(w, "article_form.html", map[string]interface{}{"Article": &models.Article{}, "IsNew": true, "CurrentUser": CurrentUser(r)})
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
		h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": true, "Error": "Nepodařilo se vytvořit článek. Ujistěte se, že slug je unikátní.", "CurrentUser": CurrentUser(r)})
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
	h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": false, "CurrentUser": CurrentUser(r)})
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
		h.render(w, "article_form.html", map[string]interface{}{"Article": article, "IsNew": false, "Error": "Nepodařilo se aktualizovat článek.", "CurrentUser": CurrentUser(r)})
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
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	for i := range galleries {
		images, _ := h.Galleries.GetImages(galleries[i].ID)
		galleries[i].Images = images
	}
	h.render(w, "galleries.html", map[string]interface{}{"Galleries": galleries, "CurrentUser": CurrentUser(r)})
}

func (h *AdminHandler) Galleries_New(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.Articles.GetAll()
	h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": &models.Gallery{}, "IsNew": true, "Articles": articles, "CurrentUser": CurrentUser(r)})
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
		h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": gallery, "IsNew": true, "Articles": articles, "Error": "Nepodařilo se vytvořit galerii. Ujistěte se, že slug je unikátní.", "CurrentUser": CurrentUser(r)})
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
	h.render(w, "gallery_form.html", map[string]interface{}{"Gallery": gallery, "IsNew": false, "Articles": articles, "CurrentUser": CurrentUser(r)})
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
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	h.render(w, "comments.html", map[string]interface{}{"Comments": comments, "CurrentUser": CurrentUser(r)})
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
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	saved := r.URL.Query().Get("saved") == "true"
	h.render(w, "settings.html", map[string]interface{}{"Settings": settings, "Saved": saved, "CurrentUser": CurrentUser(r)})
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

// --- Users (admin-only) ---

func (h *AdminHandler) Users_List(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.GetAll()
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	h.render(w, "users.html", map[string]interface{}{"Users": users, "CurrentUser": CurrentUser(r)})
}

func (h *AdminHandler) Users_New(w http.ResponseWriter, r *http.Request) {
	h.render(w, "user_form.html", map[string]interface{}{"User": &models.User{}, "IsNew": true, "CurrentUser": CurrentUser(r)})
}

func (h *AdminHandler) Users_Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	nickname := strings.TrimSpace(r.FormValue("nickname"))
	password := r.FormValue("password")
	if nickname == "" || password == "" {
		h.render(w, "user_form.html", map[string]interface{}{
			"User":        &models.User{Name: r.FormValue("name"), Surname: r.FormValue("surname"), Nickname: nickname, IsAdmin: r.FormValue("is_admin") == "on"},
			"IsNew":       true,
			"Error":       "Přezdívka a heslo jsou povinné.",
			"CurrentUser": CurrentUser(r),
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}

	user := &models.User{
		Name:         strings.TrimSpace(r.FormValue("name")),
		Surname:      strings.TrimSpace(r.FormValue("surname")),
		Nickname:     nickname,
		PasswordHash: string(hash),
		IsAdmin:      r.FormValue("is_admin") == "on",
	}

	if err := h.Users.Create(user); err != nil {
		log.Printf("error creating user: %v", err)
		h.render(w, "user_form.html", map[string]interface{}{"User": user, "IsNew": true, "Error": "Nepodařilo se vytvořit uživatele. Přezdívka může být již obsazena.", "CurrentUser": CurrentUser(r)})
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) Users_Edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user, err := h.Users.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "user_form.html", map[string]interface{}{"User": user, "IsNew": false, "CurrentUser": CurrentUser(r)})
}

func (h *AdminHandler) Users_Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user, err := h.Users.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()

	user.Name = strings.TrimSpace(r.FormValue("name"))
	user.Surname = strings.TrimSpace(r.FormValue("surname"))
	user.Nickname = strings.TrimSpace(r.FormValue("nickname"))
	user.IsAdmin = r.FormValue("is_admin") == "on"

	if password := r.FormValue("password"); password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Interní chyba serveru", 500)
			return
		}
		user.PasswordHash = string(hash)
	}

	if err := h.Users.Update(user); err != nil {
		log.Printf("error updating user: %v", err)
		h.render(w, "user_form.html", map[string]interface{}{"User": user, "IsNew": false, "Error": "Nepodařilo se aktualizovat uživatele.", "CurrentUser": CurrentUser(r)})
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) Users_Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	currentUser := CurrentUser(r)
	if currentUser != nil && currentUser.ID == id {
		http.Error(w, "Nemůžete smazat sami sebe", http.StatusBadRequest)
		return
	}
	h.Users.Delete(id)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// --- Profile (self-edit) ---

func (h *AdminHandler) Profile_Show(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r)
	saved := r.URL.Query().Get("saved") == "true"
	h.render(w, "profile.html", map[string]interface{}{"User": user, "Saved": saved, "CurrentUser": user})
}

func (h *AdminHandler) Profile_Update(w http.ResponseWriter, r *http.Request) {
	currentUser := CurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	r.ParseForm()

	user, err := h.Users.GetByID(currentUser.ID)
	if err != nil {
		http.Error(w, "Interní chyba serveru", 500)
		return
	}

	user.Name = strings.TrimSpace(r.FormValue("name"))
	user.Surname = strings.TrimSpace(r.FormValue("surname"))

	if password := r.FormValue("password"); password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Interní chyba serveru", 500)
			return
		}
		user.PasswordHash = string(hash)
	}

	if err := h.Users.Update(user); err != nil {
		log.Printf("error updating profile: %v", err)
		h.render(w, "profile.html", map[string]interface{}{"User": user, "Error": "Nepodařilo se aktualizovat profil.", "CurrentUser": user})
		return
	}

	http.Redirect(w, r, "/admin/profile?saved=true", http.StatusSeeOther)
}

func (h *AdminHandler) render(w http.ResponseWriter, name string, data interface{}) {
	t, ok := h.Templates[name]
	if !ok {
		log.Printf("admin template not found: %s", name)
		http.Error(w, "Interní chyba serveru", 500)
		return
	}
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("admin template error (%s): %v", name, err)
		http.Error(w, "Interní chyba serveru", 500)
	}
}
