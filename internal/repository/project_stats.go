package repository

import (
	"context"

	"github.com/MDYUSU/mini-jira/internal/model"
)

func (r *ProjectRepository) Stats(ctx context.Context, projectID int64) (model.ProjectStats, error) {
	s := model.ProjectStats{ProjectID: projectID}
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'todo'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'in_review'),
			COUNT(*) FILTER (WHERE status = 'done'),
			COUNT(*) FILTER (WHERE priority = 'low'),
			COUNT(*) FILTER (WHERE priority = 'medium'),
			COUNT(*) FILTER (WHERE priority = 'high'),
			COUNT(*) FILTER (WHERE priority = 'urgent'),
			COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status <> 'done'),
			COUNT(*) FILTER (WHERE assignee_id IS NULL)
		FROM tasks
		WHERE project_id = $1`, projectID).Scan(
		&s.Total,
		&s.ByStatus.Todo, &s.ByStatus.InProgress, &s.ByStatus.InReview, &s.ByStatus.Done,
		&s.ByPriority.Low, &s.ByPriority.Medium, &s.ByPriority.High, &s.ByPriority.Urgent,
		&s.Overdue, &s.Unassigned,
	)
	return s, err
}

func (r *ProjectRepository) Workload(ctx context.Context, projectID int64) ([]model.MemberWorkload, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.name, pm.role,
		       COUNT(t.id) FILTER (WHERE t.status <> 'done') AS open_tasks,
		       COUNT(t.id) FILTER (WHERE t.status = 'done'),
		       COUNT(t.id) FILTER (WHERE t.status <> 'done' AND t.priority IN ('high', 'urgent')),
		       COUNT(t.id) FILTER (WHERE t.status <> 'done' AND t.due_date < CURRENT_DATE)
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		LEFT JOIN tasks t ON t.project_id = pm.project_id AND t.assignee_id = pm.user_id
		WHERE pm.project_id = $1
		GROUP BY u.id, u.name, pm.role
		ORDER BY open_tasks DESC, u.name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.MemberWorkload{}
	for rows.Next() {
		var m model.MemberWorkload
		if err := rows.Scan(&m.UserID, &m.Name, &m.Role, &m.OpenTasks, &m.DoneTasks, &m.HighPriorityOpen, &m.Overdue); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}
