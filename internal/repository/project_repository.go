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

func (r *ProjectRepository) CreateWithOwner(ctx context.Context, orgID int64, name, key string, description *string, ownerID int64) (model.Project, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Project{}, err
	}
	defer tx.Rollback()

	var p model.Project
	err = tx.QueryRowContext(ctx, `
		INSERT INTO projects (org_id, name, key, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, key, description, created_at`, orgID, name, key, description).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Key, &p.Description, &p.CreatedAt,
	)
	if err != nil {
		return model.Project{}, mapError(err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')`, p.ID, ownerID); err != nil {
		return model.Project{}, mapError(err)
	}

	if err := tx.Commit(); err != nil {
		return model.Project{}, err
	}
	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id int64, name, key, description *string) (model.Project, error) {
	var p model.Project
	err := r.db.QueryRowContext(ctx, `
		UPDATE projects
		SET name = COALESCE($2, name),
		    key = COALESCE($3, key),
		    description = COALESCE($4, description)
		WHERE id = $1
		RETURNING id, org_id, name, key, description, created_at`, id, name, key, description).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Key, &p.Description, &p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, mapError(err)
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
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

func (r *ProjectRepository) GetMember(ctx context.Context, projectID, userID int64) (model.ProjectMember, error) {
	var m model.ProjectMember
	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.name, u.email, pm.role, pm.joined_at
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1 AND pm.user_id = $2`, projectID, userID).Scan(
		&m.UserID, &m.Name, &m.Email, &m.Role, &m.JoinedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return m, ErrNotFound
	}
	return m, err
}

func (r *ProjectRepository) AddMember(ctx context.Context, projectID, userID int64, role string) (model.ProjectMember, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)`, projectID, userID, role)
	if err != nil {
		return model.ProjectMember{}, mapError(err)
	}
	return r.GetMember(ctx, projectID, userID)
}

func (r *ProjectRepository) UpdateMemberRole(ctx context.Context, projectID, userID int64, role string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE project_members
		SET role = $3
		WHERE project_id = $1 AND user_id = $2`, projectID, userID, role)
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

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM project_members
		WHERE project_id = $1 AND user_id = $2`, projectID, userID)
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

func (r *ProjectRepository) CountOwners(ctx context.Context, projectID int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM project_members
		WHERE project_id = $1 AND role = 'owner'`, projectID).Scan(&n)
	return n, err
}