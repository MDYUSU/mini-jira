package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

type TaskListQuery struct {
	Status   string
	Priority string
	Assignee string
	Label    string
	Search   string
	TopLevel bool
	Sort     string
	Order    string
	Page     int
	Limit    int
}

var taskSortColumns = map[string]string{
	"created_at": "t.created_at",
	"due_date":   "t.due_date",
	"title":      "t.title",
	"priority":   "CASE t.priority WHEN 'low' THEN 1 WHEN 'medium' THEN 2 WHEN 'high' THEN 3 ELSE 4 END",
	"status":     "CASE t.status WHEN 'todo' THEN 1 WHEN 'in_progress' THEN 2 WHEN 'in_review' THEN 3 ELSE 4 END",
}

func (s *TaskService) ListByProjectPage(ctx context.Context, projectID int64, q TaskListQuery) (model.TaskPage, error) {
	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return model.TaskPage{}, err
	}
	if !exists {
		return model.TaskPage{}, repository.ErrNotFound
	}

	var f repository.TaskFilter

	if q.Status != "" {
		if !validStatuses[q.Status] {
			return model.TaskPage{}, ValidationError{"status must be todo, in_progress, in_review or done"}
		}
		f.Status = q.Status
	}
	if q.Priority != "" {
		if !validPriorities[q.Priority] {
			return model.TaskPage{}, ValidationError{"priority must be low, medium, high or urgent"}
		}
		f.Priority = q.Priority
	}

	switch q.Assignee {
	case "":
	case "none":
		f.Unassigned = true
	default:
		id, err := strconv.ParseInt(q.Assignee, 10, 64)
		if err != nil || id < 1 {
			return model.TaskPage{}, ValidationError{"assignee must be a user id or none"}
		}
		f.AssigneeID = &id
	}

	f.Label = strings.ToLower(strings.TrimSpace(q.Label))
	f.Search = strings.TrimSpace(q.Search)
	f.TopLevel = q.TopLevel

	sortKey := withDefault(q.Sort, "created_at")
	column, ok := taskSortColumns[sortKey]
	if !ok {
		return model.TaskPage{}, ValidationError{"sort must be created_at, due_date, priority, status or title"}
	}
	f.SortColumn = column

	order := withDefault(q.Order, "asc")
	if order != "asc" && order != "desc" {
		return model.TaskPage{}, ValidationError{"order must be asc or desc"}
	}
	f.Descending = order == "desc"

	page := q.Page
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return model.TaskPage{}, ValidationError{"page must be 1 or more"}
	}

	limit := q.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return model.TaskPage{}, ValidationError{"limit must be between 1 and 100"}
	}

	f.Limit = limit
	f.Offset = (page - 1) * limit

	tasks, total, err := s.tasks.ListByProjectFiltered(ctx, projectID, f)
	if err != nil {
		return model.TaskPage{}, err
	}

	return model.TaskPage{
		Items:      tasks,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}, nil
}