package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

type CommentInfo struct {
	TaskID    int64
	ProjectID int64
	AuthorID  *int64
}

func (r *CommentRepository) ListByTask(ctx context.Context, taskID int64) ([]model.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, u.id, u.name, c.body, c.created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.author_id
		WHERE c.task_id = $1
		ORDER BY c.created_at, c.id`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []model.Comment{}
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *CommentRepository) GetByID(ctx context.Context, id int64) (model.Comment, error) {
	c, err := scanComment(r.db.QueryRowContext(ctx, `
		SELECT c.id, u.id, u.name, c.body, c.created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.author_id
		WHERE c.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *CommentRepository) GetInfo(ctx context.Context, id int64) (CommentInfo, error) {
	var info CommentInfo
	err := r.db.QueryRowContext(ctx, `
		SELECT c.task_id, t.project_id, c.author_id
		FROM comments c
		JOIN tasks t ON t.id = c.task_id
		WHERE c.id = $1`, id).Scan(&info.TaskID, &info.ProjectID, &info.AuthorID)
	if errors.Is(err, sql.ErrNoRows) {
		return info, ErrNotFound
	}
	return info, err
}

func (r *CommentRepository) Create(ctx context.Context, taskID, authorID int64, body string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO comments (task_id, author_id, body)
		VALUES ($1, $2, $3)
		RETURNING id`, taskID, authorID, body).Scan(&id)
	return id, mapError(err)
}

func (r *CommentRepository) UpdateBody(ctx context.Context, id int64, body string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE comments
		SET body = $2, updated_at = now()
		WHERE id = $1`, id, body)
	if err != nil {
		return mapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanComment(s scanner) (model.Comment, error) {
	var c model.Comment
	var authorID *int64
	var authorName *string
	if err := s.Scan(&c.ID, &authorID, &authorName, &c.Body, &c.CreatedAt); err != nil {
		return c, err
	}
	if authorID != nil {
		c.Author = &model.UserRef{ID: *authorID, Name: *authorName}
	}
	return c, nil
}