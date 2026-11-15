package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type Credentials struct {
	UserID       int64
	PasswordHash *string
}

func (r *UserRepository) CreateWithPassword(ctx context.Context, email, name, passwordHash string) (model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, created_at, updated_at`, email, name, passwordHash).Scan(
		&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, mapError(err)
}

func (r *UserRepository) GetCredentials(ctx context.Context, email string) (Credentials, error) {
	var c Credentials
	err := r.db.QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM users
		WHERE email = $1`, email).Scan(&c.UserID, &c.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *UserRepository) GetBasic(ctx context.Context, id int64) (model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, name, created_at, updated_at
		FROM users
		WHERE id = $1`, id).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}
