package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Gallery struct {
	ID          int64
	Title       string
	Slug        string
	Description string
	ArticleID   *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Images      []Image
}

type Image struct {
	ID        int64
	GalleryID int64
	Filename  string
	Caption   string
	SortOrder int
	CreatedAt time.Time
}

type GalleryStore struct {
	DB *sql.DB
}

func (s *GalleryStore) GetAll() ([]Gallery, error) {
	rows, err := s.DB.Query("SELECT id, title, slug, description, article_id, created_at, updated_at FROM galleries ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGalleries(rows)
}

func (s *GalleryStore) GetBySlug(slug string) (*Gallery, error) {
	g := &Gallery{}
	err := s.DB.QueryRow("SELECT id, title, slug, description, article_id, created_at, updated_at FROM galleries WHERE slug = ?", slug).
		Scan(&g.ID, &g.Title, &g.Slug, &g.Description, &g.ArticleID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	images, err := s.GetImages(g.ID)
	if err != nil {
		return nil, err
	}
	g.Images = images
	return g, nil
}

func (s *GalleryStore) GetByID(id int64) (*Gallery, error) {
	g := &Gallery{}
	err := s.DB.QueryRow("SELECT id, title, slug, description, article_id, created_at, updated_at FROM galleries WHERE id = ?", id).
		Scan(&g.ID, &g.Title, &g.Slug, &g.Description, &g.ArticleID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	images, err := s.GetImages(g.ID)
	if err != nil {
		return nil, err
	}
	g.Images = images
	return g, nil
}

func (s *GalleryStore) GetByArticleID(articleID int64) (*Gallery, error) {
	g := &Gallery{}
	err := s.DB.QueryRow("SELECT id, title, slug, description, article_id, created_at, updated_at FROM galleries WHERE article_id = ?", articleID).
		Scan(&g.ID, &g.Title, &g.Slug, &g.Description, &g.ArticleID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	images, err := s.GetImages(g.ID)
	if err != nil {
		return nil, err
	}
	g.Images = images
	return g, nil
}

func (s *GalleryStore) Create(g *Gallery) error {
	res, err := s.DB.Exec("INSERT INTO galleries (title, slug, description, article_id) VALUES (?, ?, ?, ?)",
		g.Title, g.Slug, g.Description, g.ArticleID)
	if err != nil {
		return fmt.Errorf("insert gallery: %w", err)
	}
	g.ID, _ = res.LastInsertId()
	return nil
}

func (s *GalleryStore) Update(g *Gallery) error {
	_, err := s.DB.Exec("UPDATE galleries SET title=?, slug=?, description=?, article_id=? WHERE id=?",
		g.Title, g.Slug, g.Description, g.ArticleID, g.ID)
	return err
}

func (s *GalleryStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM galleries WHERE id = ?", id)
	return err
}

func (s *GalleryStore) GetImages(galleryID int64) ([]Image, error) {
	rows, err := s.DB.Query("SELECT id, gallery_id, filename, caption, sort_order, created_at FROM images WHERE gallery_id = ? ORDER BY sort_order, id", galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.ID, &img.GalleryID, &img.Filename, &img.Caption, &img.SortOrder, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (s *GalleryStore) AddImage(img *Image) error {
	res, err := s.DB.Exec("INSERT INTO images (gallery_id, filename, caption, sort_order) VALUES (?, ?, ?, ?)",
		img.GalleryID, img.Filename, img.Caption, img.SortOrder)
	if err != nil {
		return fmt.Errorf("insert image: %w", err)
	}
	img.ID, _ = res.LastInsertId()
	return nil
}

func (s *GalleryStore) DeleteImage(id int64) error {
	_, err := s.DB.Exec("DELETE FROM images WHERE id = ?", id)
	return err
}

func (s *GalleryStore) GetImageByID(id int64) (*Image, error) {
	img := &Image{}
	err := s.DB.QueryRow("SELECT id, gallery_id, filename, caption, sort_order, created_at FROM images WHERE id = ?", id).
		Scan(&img.ID, &img.GalleryID, &img.Filename, &img.Caption, &img.SortOrder, &img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func scanGalleries(rows *sql.Rows) ([]Gallery, error) {
	var galleries []Gallery
	for rows.Next() {
		var g Gallery
		if err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Description, &g.ArticleID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		galleries = append(galleries, g)
	}
	return galleries, rows.Err()
}
