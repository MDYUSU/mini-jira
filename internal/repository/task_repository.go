package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

type NewTask struct {
	ProjectID    int64
	ParentTaskID *int64
	Title        string
	Description  *string
	Status       string
	Priority     string
	AssigneeID   *int64
	ReporterID   int64
	DueDate      *time.Time
}

type TaskChanges struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	SetAssignee bool
	AssigneeID  *int64
	SetDueDate  bool
	DueDate     *time.Time
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

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (model.Task, error) {
	t, err := scanTask(r.db.QueryRowContext(ctx, taskSelect+` WHERE t.id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (r *TaskRepository) GetProjectID(ctx context.Context, id int64) (int64, error) {
	var projectID int64
	err := r.db.QueryRowContext(ctx, `SELECT project_id FROM tasks WHERE id = $1`, id).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return projectID, err
}

func (r *TaskRepository) GetDetail(ctx context.Context, id int64) (model.TaskDetail, error) {
	t, err := r.GetByID(ctx, id)
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
		ORDER BY a.created_at, a.id`, taskID)
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

func (r *TaskRepository) Create(ctx context.Context, t NewTask) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tasks (project_id, parent_task_id, title, description, status, priority, assignee_id, reporter_id, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		t.ProjectID, t.ParentTaskID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.ReporterID, t.DueDate,
	).Scan(&id)
	if err != nil {
		return 0, mapError(err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs (task_id, actor_id, action)
		VALUES ($1, $2, 'created')`, id, t.ReporterID); err != nil {
		return 0, mapError(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

type activityEntry struct {
	action   string
	field    string
	oldValue *string
	newValue *string
}

func (r *TaskRepository) Update(ctx context.Context, id, actorID int64, c TaskChanges) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var cur struct {
		title       string
		description *string
		status      string
		priority    string
		assigneeID  *int64
		dueDate     *time.Time
	}
	err = tx.QueryRowContext(ctx, `
		SELECT title, description, status, priority, assignee_id, due_date
		FROM tasks
		WHERE id = $1
		FOR UPDATE`, id).Scan(&cur.title, &cur.description, &cur.status, &cur.priority, &cur.assigneeID, &cur.dueDate)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	next := cur
	var logs []activityEntry

	if c.Title != nil && *c.Title != cur.title {
		logs = append(logs, activityEntry{"title_changed", "title", &cur.title, c.Title})
		next.title = *c.Title
	}
	if c.Description != nil && !equalString(c.Description, cur.description) {
		logs = append(logs, activityEntry{"description_changed", "description", cur.description, c.Description})
		next.description = c.Description
	}
	if c.Status != nil && *c.Status != cur.status {
		logs = append(logs, activityEntry{"status_changed", "status", &cur.status, c.Status})
		next.status = *c.Status
	}
	if c.Priority != nil && *c.Priority != cur.priority {
		logs = append(logs, activityEntry{"priority_changed", "priority", &cur.priority, c.Priority})
		next.priority = *c.Priority
	}
	if c.SetAssignee && !equalInt(c.AssigneeID, cur.assigneeID) {
		action := "assigned"
		if c.AssigneeID == nil {
			action = "unassigned"
		}
		logs = append(logs, activityEntry{action, "assignee_id", intText(cur.assigneeID), intText(c.AssigneeID)})
		next.assigneeID = c.AssigneeID
	}
	if c.SetDueDate && !equalDate(c.DueDate, cur.dueDate) {
		logs = append(logs, activityEntry{"due_date_changed", "due_date", dateText(cur.dueDate), dateText(c.DueDate)})
		next.dueDate = c.DueDate
	}

	if len(logs) == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tasks
		SET title = $2, description = $3, status = $4, priority = $5,
		    assignee_id = $6, due_date = $7, updated_at = now()
		WHERE id = $1`,
		id, next.title, next.description, next.status, next.priority, next.assigneeID, next.dueDate,
	); err != nil {
		return mapError(err)
	}

	for _, l := range logs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO activity_logs (task_id, actor_id, action, field, old_value, new_value)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			id, actorID, l.action, l.field, l.oldValue, l.newValue,
		); err != nil {
			return mapError(err)
		}
	}

	return tx.Commit()
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func equalString(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func equalInt(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func equalDate(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func intText(v *int64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}

func dateText(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.Format("2006-01-02")
	return &s
}