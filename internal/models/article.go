package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Article struct {
	ID         int64
	Title      string
	Slug       string
	Content    string
	Excerpt    string
	CoverImage string
	Published  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ArticleStore struct {
	DB *sql.DB
}

func (s *ArticleStore) GetAll() ([]Article, error) {
	rows, err := s.DB.Query("SELECT id, title, slug, content, excerpt, cover_image, published, created_at, updated_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *ArticleStore) GetPublished() ([]Article, error) {
	rows, err := s.DB.Query("SELECT id, title, slug, content, excerpt, cover_image, published, created_at, updated_at FROM articles WHERE published = TRUE ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArticles(rows)
}

func (s *ArticleStore) GetPublishedPaginated(limit, offset int) ([]Article, int, error) {
	var total int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM articles WHERE published = TRUE").Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.DB.Query("SELECT id, title, slug, content, excerpt, cover_image, published, created_at, updated_at FROM articles WHERE published = TRUE ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	articles, err := scanArticles(rows)
	return articles, total, err
}

func (s *ArticleStore) GetBySlug(slug string) (*Article, error) {
	a := &Article{}
	err := s.DB.QueryRow("SELECT id, title, slug, content, excerpt, cover_image, published, created_at, updated_at FROM articles WHERE slug = ?", slug).
		Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Published, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ArticleStore) GetByID(id int64) (*Article, error) {
	a := &Article{}
	err := s.DB.QueryRow("SELECT id, title, slug, content, excerpt, cover_image, published, created_at, updated_at FROM articles WHERE id = ?", id).
		Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Published, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ArticleStore) Create(a *Article) error {
	res, err := s.DB.Exec("INSERT INTO articles (title, slug, content, excerpt, cover_image, published) VALUES (?, ?, ?, ?, ?, ?)",
		a.Title, a.Slug, a.Content, a.Excerpt, a.CoverImage, a.Published)
	if err != nil {
		return fmt.Errorf("insert article: %w", err)
	}
	a.ID, _ = res.LastInsertId()
	return nil
}

func (s *ArticleStore) Update(a *Article) error {
	_, err := s.DB.Exec("UPDATE articles SET title=?, slug=?, content=?, excerpt=?, cover_image=?, published=? WHERE id=?",
		a.Title, a.Slug, a.Content, a.Excerpt, a.CoverImage, a.Published, a.ID)
	return err
}

func (s *ArticleStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM articles WHERE id = ?", id)
	return err
}

func scanArticles(rows *sql.Rows) ([]Article, error) {
	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Content, &a.Excerpt, &a.CoverImage, &a.Published, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}
