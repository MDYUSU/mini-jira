package model

import "time"

type UserRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID           int64      `json:"id"`
	ProjectID    int64      `json:"project_id"`
	ParentTaskID *int64     `json:"parent_task_id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority"`
	Assignee     *UserRef   `json:"assignee"`
	Reporter     UserRef    `json:"reporter"`
	DueDate      *time.Time `json:"due_date"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type TaskDetail struct {
	Task
	Labels   []Label   `json:"labels"`
	Subtasks []Task    `json:"subtasks"`
	Comments []Comment `json:"comments"`
}