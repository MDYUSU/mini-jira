package model

import "time"

type ActivityLog struct {
	ID        int64     `json:"id"`
	Actor     *UserRef  `json:"actor"`
	Action    string    `json:"action"`
	Field     *string   `json:"field"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}