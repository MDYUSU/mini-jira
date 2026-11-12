package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

const taskSelect = `
	SELECT t.id, t.project_id, t.parent_task_id, t.title, t.description, t.status, t.priority,
	       a.id, a.name, r.id, r.name, t.due_date, t.created_at, t.updated_at
	FROM tasks t
	LEFT JOIN users a ON a.id = t.assignee_id
	JOIN users r ON r.id = t.reporter_id`

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (model.Task, error) {
	var t model.Task
	var assigneeID *int64
	var assigneeName *string

	err := s.Scan(
		&t.ID, &t.ProjectID, &t.ParentTaskID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&assigneeID, &assigneeName, &t.Reporter.ID, &t.Reporter.Name,
		&t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return t, err
	}

	if assigneeID != nil {
		t.Assignee = &model.UserRef{ID: *assigneeID, Name: *assigneeName}
	}
	return t, nil
}

func (r *TaskRepository) queryTasks(ctx context.Context, query string, args ...any) ([]model.Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) ListByProject(ctx context.Context, projectID int64) ([]model.Task, error) {
	return r.queryTasks(ctx, taskSelect+` WHERE t.project_id = $1 ORDER BY t.id`, projectID)
}

func (r *TaskRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM tasks WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *TaskRepository) GetDetail(ctx context.Context, id int64) (model.TaskDetail, error) {
	t, err := scanTask(r.db.QueryRowContext(ctx, taskSelect+` WHERE t.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.TaskDetail{}, ErrNotFound
	}
	if err != nil {
		return model.TaskDetail{}, err
	}

	d := model.TaskDetail{Task: t}

	d.Labels, err = r.listLabels(ctx, id)
	if err != nil {
		return d, err
	}

	d.Subtasks, err = r.queryTasks(ctx, taskSelect+` WHERE t.parent_task_id = $1 ORDER BY t.id`, id)
	if err != nil {
		return d, err
	}

	d.Comments, err = r.listComments(ctx, id)
	if err != nil {
		return d, err
	}

	return d, nil
}

func (r *TaskRepository) listLabels(ctx context.Context, taskID int64) ([]model.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.name, l.color
		FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id
		WHERE tl.task_id = $1
		ORDER BY l.name`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	labels := []model.Label{}
	for rows.Next() {
		var l model.Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (r *TaskRepository) listComments(ctx context.Context, taskID int64) ([]model.Comment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, u.id, u.name, c.body, c.created_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.author_id
		WHERE c.task_id = $1
		ORDER BY c.created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []model.Comment{}
	for rows.Next() {
		var c model.Comment
		var authorID *int64
		var authorName *string
		if err := rows.Scan(&c.ID, &authorID, &authorName, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		if authorID != nil {
			c.Author = &model.UserRef{ID: *authorID, Name: *authorName}
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (r *TaskRepository) ListActivity(ctx context.Context, taskID int64) ([]model.ActivityLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, u.id, u.name, a.action, a.field, a.old_value, a.new_value, a.created_at
		FROM activity_logs a
		LEFT JOIN users u ON u.id = a.actor_id
		WHERE a.task_id = $1
		ORDER BY a.created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []model.ActivityLog{}
	for rows.Next() {
		var l model.ActivityLog
		var actorID *int64
		var actorName *string
		if err := rows.Scan(&l.ID, &actorID, &actorName, &l.Action, &l.Field, &l.OldValue, &l.NewValue, &l.CreatedAt); err != nil {
			return nil, err
		}
		if actorID != nil {
			l.Actor = &model.UserRef{ID: *actorID, Name: *actorName}
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}