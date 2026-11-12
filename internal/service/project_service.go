package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

var ErrLastOwner = errors.New("project must have at least one owner")

var projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

var validRoles = map[string]bool{
	"owner":  true,
	"member": true,
	"viewer": true,
}

type CreateProjectInput struct {
	Name        string  `json:"name"`
	Key         string  `json:"key"`
	Description *string `json:"description"`
	OwnerID     int64   `json:"owner_id"`
}

type UpdateProjectInput struct {
	Name        *string `json:"name"`
	Key         *string `json:"key"`
	Description *string `json:"description"`
}

type AddMemberInput struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type UpdateMemberInput struct {
	Role string `json:"role"`
}

type ProjectService struct {
	projects *repository.ProjectRepository
	orgs     *repository.OrganizationRepository
}

func NewProjectService(projects *repository.ProjectRepository, orgs *repository.OrganizationRepository) *ProjectService {
	return &ProjectService{projects: projects, orgs: orgs}
}

func (s *ProjectService) Get(ctx context.Context, id int64) (model.ProjectDetail, error) {
	return s.projects.GetDetail(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, orgID int64, in CreateProjectInput) (model.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.Project{}, ValidationError{"name is required"}
	}
	key, err := normalizeKey(in.Key)
	if err != nil {
		return model.Project{}, err
	}
	if in.OwnerID < 1 {
		return model.Project{}, ValidationError{"owner_id is required"}
	}

	if _, err := s.orgs.GetByID(ctx, orgID); err != nil {
		return model.Project{}, err
	}

	p, err := s.projects.CreateWithOwner(ctx, orgID, name, key, in.Description, in.OwnerID)
	if errors.Is(err, repository.ErrForeignKey) {
		return p, ValidationError{"owner does not exist"}
	}
	return p, err
}

func (s *ProjectService) Update(ctx context.Context, id int64, in UpdateProjectInput) (model.Project, error) {
	if in.Name == nil && in.Key == nil && in.Description == nil {
		return model.Project{}, ValidationError{"nothing to update"}
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return model.Project{}, ValidationError{"name cannot be empty"}
		}
		in.Name = &name
	}
	if in.Key != nil {
		key, err := normalizeKey(*in.Key)
		if err != nil {
			return model.Project{}, err
		}
		in.Key = &key
	}
	return s.projects.Update(ctx, id, in.Name, in.Key, in.Description)
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	return s.projects.Delete(ctx, id)
}

func (s *ProjectService) AddMember(ctx context.Context, projectID int64, in AddMemberInput) (model.ProjectMember, error) {
	if in.UserID < 1 {
		return model.ProjectMember{}, ValidationError{"user_id is required"}
	}
	role := strings.TrimSpace(in.Role)
	if role == "" {
		role = "member"
	}
	if !validRoles[role] {
		return model.ProjectMember{}, ValidationError{"role must be owner, member or viewer"}
	}

	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return model.ProjectMember{}, err
	}
	if !exists {
		return model.ProjectMember{}, repository.ErrNotFound
	}

	m, err := s.projects.AddMember(ctx, projectID, in.UserID, role)
	if errors.Is(err, repository.ErrForeignKey) {
		return m, ValidationError{"user does not exist"}
	}
	return m, err
}

func (s *ProjectService) UpdateMemberRole(ctx context.Context, projectID, userID int64, in UpdateMemberInput) (model.ProjectMember, error) {
	role := strings.TrimSpace(in.Role)
	if !validRoles[role] {
		return model.ProjectMember{}, ValidationError{"role must be owner, member or viewer"}
	}

	current, err := s.projects.GetMember(ctx, projectID, userID)
	if err != nil {
		return model.ProjectMember{}, err
	}
	if current.Role == "owner" && role != "owner" {
		if err := s.ensureAnotherOwner(ctx, projectID); err != nil {
			return model.ProjectMember{}, err
		}
	}

	if err := s.projects.UpdateMemberRole(ctx, projectID, userID, role); err != nil {
		return model.ProjectMember{}, err
	}
	return s.projects.GetMember(ctx, projectID, userID)
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID int64) error {
	current, err := s.projects.GetMember(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if current.Role == "owner" {
		if err := s.ensureAnotherOwner(ctx, projectID); err != nil {
			return err
		}
	}
	return s.projects.RemoveMember(ctx, projectID, userID)
}

func (s *ProjectService) ensureAnotherOwner(ctx context.Context, projectID int64) error {
	n, err := s.projects.CountOwners(ctx, projectID)
	if err != nil {
		return err
	}
	if n <= 1 {
		return ErrLastOwner
	}
	return nil
}

func normalizeKey(raw string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(raw))
	if !projectKeyPattern.MatchString(key) {
		return "", ValidationError{"key must be 2-10 letters or digits, starting with a letter"}
	}
	return key, nil
}