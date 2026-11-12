package model

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	Author    *UserRef  `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}