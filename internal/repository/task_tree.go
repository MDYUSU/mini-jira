package repository

import (
	"context"

	"github.com/MDYUSU/mini-jira/internal/model"
)

const subtreeQuery = `
	WITH RECURSIVE tree AS (
		SELECT id, ARRAY[id] AS path
		FROM tasks
		WHERE id = $1

		UNION ALL

		SELECT t.id, tree.path || t.id
		FROM tasks t
		JOIN tree ON t.parent_task_id = tree.id
		WHERE NOT t.id = ANY(tree.path)
	)
	SELECT t.id, t.project_id, t.parent_task_id, t.title, t.description, t.status, t.priority,
	       a.id, a.name, r.id, r.name, t.due_date, t.created_at, t.updated_at
	FROM tree
	JOIN tasks t ON t.id = tree.id
	LEFT JOIN users a ON a.id = t.assignee_id
	JOIN users r ON r.id = t.reporter_id
	ORDER BY tree.path`

func (r *TaskRepository) ListSubtree(ctx context.Context, rootID int64) ([]model.Task, error) {
	return r.queryTasks(ctx, subtreeQuery, rootID)
}