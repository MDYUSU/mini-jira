package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

var ErrLabelNotAttached = errors.New("label is not attached to this task")

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type CreateLabelInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AttachLabelInput struct {
	LabelID int64 `json:"label_id"`
	ActorID int64 `json:"actor_id"`
}

type LabelService struct {
	labels   *repository.LabelRepository
	tasks    *repository.TaskRepository
	projects *repository.ProjectRepository
}

func NewLabelService(labels *repository.LabelRepository, tasks *repository.TaskRepository, projects *repository.ProjectRepository) *LabelService {
	return &LabelService{labels: labels, tasks: tasks, projects: projects}
}

func (s *LabelService) ListByProject(ctx context.Context, projectID int64) ([]model.Label, error) {
	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}
	return s.labels.ListByProject(ctx, projectID)
}

func (s *LabelService) Create(ctx context.Context, projectID int64, in CreateLabelInput) (model.Label, error) {
	name := strings.ToLower(strings.TrimSpace(in.Name))
	if name == "" {
		return model.Label{}, ValidationError{"name is required"}
	}
	color := withDefault(in.Color, "#888888")
	if !colorPattern.MatchString(color) {
		return model.Label{}, ValidationError{"color must look like #RRGGBB, for example #2563eb"}
	}

	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return model.Label{}, err
	}
	if !exists {
		return model.Label{}, repository.ErrNotFound
	}

	return s.labels.Create(ctx, projectID, name, color)
}

func (s *LabelService) Delete(ctx context.Context, id int64) error {
	return s.labels.Delete(ctx, id)
}

func (s *LabelService) Attach(ctx context.Context, taskID int64, in AttachLabelInput) ([]model.Label, error) {
	if in.LabelID < 1 {
		return nil, ValidationError{"label_id is required"}
	}
	if in.ActorID < 1 {
		return nil, ValidationError{"actor_id is required"}
	}

	taskProject, err := s.tasks.GetProjectID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	labelProject, err := s.labels.GetProjectID(ctx, in.LabelID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ValidationError{"label does not exist"}
	}
	if err != nil {
		return nil, err
	}
	if labelProject != taskProject {
		return nil, ValidationError{"label belongs to a different project"}
	}

	if err := checkEditor(ctx, s.projects, taskProject, in.ActorID, "actor"); err != nil {
		return nil, err
	}

	if err := s.labels.AttachToTask(ctx, taskID, in.LabelID, in.ActorID); err != nil {
		return nil, err
	}

	detail, err := s.tasks.GetDetail(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return detail.Labels, nil
}

func (s *LabelService) Detach(ctx context.Context, taskID, labelID, actorID int64) error {
	taskProject, err := s.tasks.GetProjectID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := checkEditor(ctx, s.projects, taskProject, actorID, "actor"); err != nil {
		return err
	}

	err = s.labels.DetachFromTask(ctx, taskID, labelID, actorID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrLabelNotAttached
	}
	return err
}