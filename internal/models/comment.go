package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Comment struct {
	ID         int64
	ArticleID  int64
	AuthorName string
	Content    string
	Approved   bool
	CreatedAt  time.Time
}

type CommentStore struct {
	DB *sql.DB
}

func (s *CommentStore) GetByArticleID(articleID int64, approvedOnly bool) ([]Comment, error) {
	query := "SELECT id, article_id, author_name, content, approved, created_at FROM comments WHERE article_id = ?"
	if approvedOnly {
		query += " AND approved = TRUE"
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.DB.Query(query, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComments(rows)
}

func (s *CommentStore) GetAllPending() ([]Comment, error) {
	rows, err := s.DB.Query("SELECT id, article_id, author_name, content, approved, created_at FROM comments WHERE approved = FALSE ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComments(rows)
}

func (s *CommentStore) GetAll() ([]Comment, error) {
	rows, err := s.DB.Query("SELECT id, article_id, author_name, content, approved, created_at FROM comments ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanComments(rows)
}

func (s *CommentStore) Create(c *Comment) error {
	res, err := s.DB.Exec("INSERT INTO comments (article_id, author_name, content, approved) VALUES (?, ?, ?, ?)",
		c.ArticleID, c.AuthorName, c.Content, c.Approved)
	if err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (s *CommentStore) Approve(id int64) error {
	_, err := s.DB.Exec("UPDATE comments SET approved = TRUE WHERE id = ?", id)
	return err
}

func (s *CommentStore) Delete(id int64) error {
	_, err := s.DB.Exec("DELETE FROM comments WHERE id = ?", id)
	return err
}

func scanComments(rows *sql.Rows) ([]Comment, error) {
	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.ArticleID, &c.AuthorName, &c.Content, &c.Approved, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
