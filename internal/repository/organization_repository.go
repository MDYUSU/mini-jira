package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id int64) (model.Organization, error) {
	var o model.Organization
	err := r.db.QueryRowContext(ctx, `
		SELECT o.id, o.name, o.owner_id, u.name, o.created_at
		FROM organizations o
		JOIN users u ON u.id = o.owner_id
		WHERE o.id = $1`, id).Scan(&o.ID, &o.Name, &o.OwnerID, &o.OwnerName, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return o, ErrNotFound
	}
	return o, err
}

func (r *OrganizationRepository) Create(ctx context.Context, name string, ownerID int64) (model.Organization, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO organizations (name, owner_id)
		VALUES ($1, $2)
		RETURNING id`, name, ownerID).Scan(&id)
	if err != nil {
		return model.Organization{}, mapError(err)
	}
	return r.GetByID(ctx, id)
}