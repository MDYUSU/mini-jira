package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MDYUSU/mini-jira/internal/model"
)


type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, name, created_at, updated_at
		FROM users
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.UserWithProfile, error) {
	var u model.UserWithProfile
	var profileUserID *int64
	var bio, avatarURL *string
	var profileUpdatedAt *time.Time

	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.email, u.name, u.created_at, u.updated_at,
		       p.user_id, p.bio, p.avatar_url, p.updated_at
		FROM users u
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1`, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt,
		&profileUserID, &bio, &avatarURL, &profileUpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	if err != nil {
		return u, err
	}

	if profileUserID != nil {
		u.Profile = &model.UserProfile{
			UserID:    *profileUserID,
			Bio:       bio,
			AvatarURL: avatarURL,
			UpdatedAt: *profileUpdatedAt,
		}
	}
	return u, nil
}