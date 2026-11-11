package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserProfile struct {
	UserID    int64     `json:"user_id"`
	Bio       *string   `json:"bio"`
	AvatarURL *string   `json:"avatar_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserWithProfile struct {
	User
	Profile *UserProfile `json:"profile"`
}