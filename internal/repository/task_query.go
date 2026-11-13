package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type TaskFilter struct {
	Status     string
	Priority   string
	AssigneeID *int64
	Unassigned bool
	Label      string
	Search     string
	TopLevel   bool
	SortColumn string
	Descending bool
	Limit      int
	Offset     int
}

func (r *TaskRepository) ListByProjectFiltered(ctx context.Context, projectID int64, f TaskFilter) ([]model.Task, int, error) {
	where := []string{"t.project_id = $1"}
	args := []any{projectID}

	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}

	if f.Status != "" {
		add("t.status = $%d", f.Status)
	}
	if f.Priority != "" {
		add("t.priority = $%d", f.Priority)
	}
	if f.AssigneeID != nil {
		add("t.assignee_id = $%d", *f.AssigneeID)
	}
	if f.Unassigned {
		where = append(where, "t.assignee_id IS NULL")
	}
	if f.Label != "" {
		add(`EXISTS (
			SELECT 1 FROM task_labels tl
			JOIN labels l ON l.id = tl.label_id
			WHERE tl.task_id = t.id AND l.name = $%d)`, f.Label)
	}
	if f.Search != "" {
		add("t.title ILIKE $%d", "%"+f.Search+"%")
	}
	if f.TopLevel {
		where = append(where, "t.parent_task_id IS NULL")
	}

	whereSQL := " WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks t`+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	direction := "ASC"
	if f.Descending {
		direction = "DESC"
	}
	orderSQL := fmt.Sprintf(" ORDER BY %s %s NULLS LAST, t.id %s", f.SortColumn, direction, direction)

	pageArgs := append(append([]any{}, args...), f.Limit, f.Offset)
	limitSQL := fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	tasks, err := r.queryTasks(ctx, taskSelect+whereSQL+orderSQL+limitSQL, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}