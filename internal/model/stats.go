package model

type StatusCounts struct {
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	InReview   int `json:"in_review"`
	Done       int `json:"done"`
}

type PriorityCounts struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
	Urgent int `json:"urgent"`
}

type ProjectStats struct {
	ProjectID   int64          `json:"project_id"`
	Total       int            `json:"total"`
	ByStatus    StatusCounts   `json:"by_status"`
	ByPriority  PriorityCounts `json:"by_priority"`
	Overdue     int            `json:"overdue"`
	Unassigned  int            `json:"unassigned"`
	DonePercent float64        `json:"done_percent"`
}

type MemberWorkload struct {
	UserID           int64  `json:"user_id"`
	Name             string `json:"name"`
	Role             string `json:"role"`
	OpenTasks        int    `json:"open_tasks"`
	DoneTasks        int    `json:"done_tasks"`
	HighPriorityOpen int    `json:"high_priority_open"`
	Overdue          int    `json:"overdue"`
}
