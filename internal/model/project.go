package model

import "time"

type Project struct {
	ID          int64     `json:"id"`
	OrgID       int64     `json:"org_id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectMember struct {
	UserID   int64     `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type ProjectDetail struct {
	Project
	OrgName string          `json:"org_name"`
	Members []ProjectMember `json:"members"`
}