package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) ListByOrg(ctx context.Context, orgID int64) ([]model.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, org_id, name, key, description, created_at
		FROM projects
		WHERE org_id = $1
		ORDER BY id`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []model.Project{}
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Key, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) GetDetail(ctx context.Context, id int64) (model.ProjectDetail, error) {
	var d model.ProjectDetail
	err := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.org_id, p.name, p.key, p.description, p.created_at, o.name
		FROM projects p
		JOIN organizations o ON o.id = p.org_id
		WHERE p.id = $1`, id).Scan(
		&d.ID, &d.OrgID, &d.Name, &d.Key, &d.Description, &d.CreatedAt, &d.OrgName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.name, u.email, pm.role, pm.joined_at
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY u.name`, id)
	if err != nil {
		return d, err
	}
	defer rows.Close()

	d.Members = []model.ProjectMember{}
	for rows.Next() {
		var m model.ProjectMember
		if err := rows.Scan(&m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return d, err
		}
		d.Members = append(d.Members, m)
	}
	return d, rows.Err()
}

func (r *ProjectRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM projects WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}