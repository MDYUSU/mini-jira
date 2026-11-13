package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type LabelRepository struct {
	db *sql.DB
}

func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

func (r *LabelRepository) ListByProject(ctx context.Context, projectID int64) ([]model.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, color
		FROM labels
		WHERE project_id = $1
		ORDER BY name`, projectID)
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

func (r *LabelRepository) Create(ctx context.Context, projectID int64, name, color string) (model.Label, error) {
	var l model.Label
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO labels (project_id, name, color)
		VALUES ($1, $2, $3)
		RETURNING id, name, color`, projectID, name, color).Scan(&l.ID, &l.Name, &l.Color)
	return l, mapError(err)
}

func (r *LabelRepository) GetProjectID(ctx context.Context, id int64) (int64, error) {
	var projectID int64
	err := r.db.QueryRowContext(ctx, `SELECT project_id FROM labels WHERE id = $1`, id).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return projectID, err
}

func (r *LabelRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM labels WHERE id = $1`, id)
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

func (r *LabelRepository) AttachToTask(ctx context.Context, taskID, labelID, actorID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name string
	err = tx.QueryRowContext(ctx, `SELECT name FROM labels WHERE id = $1`, labelID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO task_labels (task_id, label_id)
		VALUES ($1, $2)`, taskID, labelID); err != nil {
		return mapError(err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs (task_id, actor_id, action, field, new_value)
		VALUES ($1, $2, 'label_added', 'labels', $3)`, taskID, actorID, name); err != nil {
		return mapError(err)
	}

	return tx.Commit()
}

func (r *LabelRepository) DetachFromTask(ctx context.Context, taskID, labelID, actorID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name string
	err = tx.QueryRowContext(ctx, `
		SELECT l.name
		FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id
		WHERE tl.task_id = $1 AND tl.label_id = $2`, taskID, labelID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM task_labels
		WHERE task_id = $1 AND label_id = $2`, taskID, labelID); err != nil {
		return mapError(err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs (task_id, actor_id, action, field, old_value)
		VALUES ($1, $2, 'label_removed', 'labels', $3)`, taskID, actorID, name); err != nil {
		return mapError(err)
	}

	return tx.Commit()
}