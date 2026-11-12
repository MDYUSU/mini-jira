package service

import (
	"context"
	"errors"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

type CreateOrganizationInput struct {
	Name    string `json:"name"`
	OwnerID int64  `json:"owner_id"`
}

type OrganizationService struct {
	orgs     *repository.OrganizationRepository
	projects *repository.ProjectRepository
}

func NewOrganizationService(orgs *repository.OrganizationRepository, projects *repository.ProjectRepository) *OrganizationService {
	return &OrganizationService{orgs: orgs, projects: projects}
}

func (s *OrganizationService) Get(ctx context.Context, id int64) (model.Organization, error) {
	return s.orgs.GetByID(ctx, id)
}

func (s *OrganizationService) ListProjects(ctx context.Context, orgID int64) ([]model.Project, error) {
	if _, err := s.orgs.GetByID(ctx, orgID); err != nil {
		return nil, err
	}
	return s.projects.ListByOrg(ctx, orgID)
}

func (s *OrganizationService) Create(ctx context.Context, in CreateOrganizationInput) (model.Organization, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.Organization{}, ValidationError{"name is required"}
	}
	if in.OwnerID < 1 {
		return model.Organization{}, ValidationError{"owner_id is required"}
	}
	org, err := s.orgs.Create(ctx, name, in.OwnerID)
	if errors.Is(err, repository.ErrForeignKey) {
		return org, ValidationError{"owner does not exist"}
	}
	return org, err
}