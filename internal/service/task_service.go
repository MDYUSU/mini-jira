package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

var ErrViewerReadOnly = errors.New("viewers cannot create or change tasks")

var validStatuses = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"in_review":   true,
	"done":        true,
}

var validPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

type CreateTaskInput struct {
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	Priority     string  `json:"priority"`
	AssigneeID   *int64  `json:"assignee_id"`
	ParentTaskID *int64  `json:"parent_task_id"`
	DueDate      *string `json:"due_date"`
	ReporterID   int64   `json:"reporter_id"`
}

type UpdateTaskInput struct {
	ActorID     int64   `json:"actor_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *int64  `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type TaskService struct {
	tasks    *repository.TaskRepository
	projects *repository.ProjectRepository
}

func NewTaskService(tasks *repository.TaskRepository, projects *repository.ProjectRepository) *TaskService {
	return &TaskService{tasks: tasks, projects: projects}
}

func (s *TaskService) ListByProject(ctx context.Context, projectID int64) ([]model.Task, error) {
	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}
	return s.tasks.ListByProject(ctx, projectID)
}

func (s *TaskService) Get(ctx context.Context, id int64) (model.TaskDetail, error) {
	return s.tasks.GetDetail(ctx, id)
}

func (s *TaskService) ListActivity(ctx context.Context, id int64) ([]model.ActivityLog, error) {
	exists, err := s.tasks.Exists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}
	return s.tasks.ListActivity(ctx, id)
}

func (s *TaskService) Create(ctx context.Context, projectID int64, in CreateTaskInput) (model.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return model.Task{}, ValidationError{"title is required"}
	}
	status := withDefault(in.Status, "todo")
	if !validStatuses[status] {
		return model.Task{}, ValidationError{"status must be todo, in_progress, in_review or done"}
	}
	priority := withDefault(in.Priority, "medium")
	if !validPriorities[priority] {
		return model.Task{}, ValidationError{"priority must be low, medium, high or urgent"}
	}
	if in.ReporterID < 1 {
		return model.Task{}, ValidationError{"reporter_id is required"}
	}
	dueDate, err := parseDate(in.DueDate)
	if err != nil {
		return model.Task{}, err
	}

	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return model.Task{}, err
	}
	if !exists {
		return model.Task{}, repository.ErrNotFound
	}

	if err := s.requireEditor(ctx, projectID, in.ReporterID, "reporter"); err != nil {
		return model.Task{}, err
	}

	if in.AssigneeID != nil {
		if *in.AssigneeID < 1 {
			return model.Task{}, ValidationError{"assignee_id must be a positive number"}
		}
		if err := s.requireMember(ctx, projectID, *in.AssigneeID, "assignee"); err != nil {
			return model.Task{}, err
		}
	}

	if in.ParentTaskID != nil {
		parentProject, err := s.tasks.GetProjectID(ctx, *in.ParentTaskID)
		if errors.Is(err, repository.ErrNotFound) {
			return model.Task{}, ValidationError{"parent task does not exist"}
		}
		if err != nil {
			return model.Task{}, err
		}
		if parentProject != projectID {
			return model.Task{}, ValidationError{"parent task belongs to a different project"}
		}
	}

	id, err := s.tasks.Create(ctx, repository.NewTask{
		ProjectID:    projectID,
		ParentTaskID: in.ParentTaskID,
		Title:        title,
		Description:  in.Description,
		Status:       status,
		Priority:     priority,
		AssigneeID:   in.AssigneeID,
		ReporterID:   in.ReporterID,
		DueDate:      dueDate,
	})
	if err != nil {
		return model.Task{}, err
	}
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) Update(ctx context.Context, id int64, in UpdateTaskInput) (model.Task, error) {
	if in.ActorID < 1 {
		return model.Task{}, ValidationError{"actor_id is required"}
	}
	if in.Title == nil && in.Description == nil && in.Status == nil && in.Priority == nil && in.AssigneeID == nil && in.DueDate == nil {
		return model.Task{}, ValidationError{"nothing to update"}
	}

	projectID, err := s.tasks.GetProjectID(ctx, id)
	if err != nil {
		return model.Task{}, err
	}
	if err := s.requireEditor(ctx, projectID, in.ActorID, "actor"); err != nil {
		return model.Task{}, err
	}

	var changes repository.TaskChanges

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return model.Task{}, ValidationError{"title cannot be empty"}
		}
		changes.Title = &title
	}
	changes.Description = in.Description

	if in.Status != nil {
		status := strings.TrimSpace(*in.Status)
		if !validStatuses[status] {
			return model.Task{}, ValidationError{"status must be todo, in_progress, in_review or done"}
		}
		changes.Status = &status
	}
	if in.Priority != nil {
		priority := strings.TrimSpace(*in.Priority)
		if !validPriorities[priority] {
			return model.Task{}, ValidationError{"priority must be low, medium, high or urgent"}
		}
		changes.Priority = &priority
	}

	if in.AssigneeID != nil {
		changes.SetAssignee = true
		switch {
		case *in.AssigneeID < 0:
			return model.Task{}, ValidationError{"assignee_id must be 0 (unassign) or a user id"}
		case *in.AssigneeID > 0:
			if err := s.requireMember(ctx, projectID, *in.AssigneeID, "assignee"); err != nil {
				return model.Task{}, err
			}
			changes.AssigneeID = in.AssigneeID
		}
	}

	if in.DueDate != nil {
		dueDate, err := parseDate(in.DueDate)
		if err != nil {
			return model.Task{}, err
		}
		changes.SetDueDate = true
		changes.DueDate = dueDate
	}

	if err := s.tasks.Update(ctx, id, in.ActorID, changes); err != nil {
		return model.Task{}, err
	}
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.tasks.Delete(ctx, id)
}

func (s *TaskService) requireMember(ctx context.Context, projectID, userID int64, who string) error {
	_, err := s.projects.GetMember(ctx, projectID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ValidationError{who + " must be a member of this project"}
	}
	return err
}

func (s *TaskService) requireEditor(ctx context.Context, projectID, userID int64, who string) error {
	m, err := s.projects.GetMember(ctx, projectID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ValidationError{who + " must be a member of this project"}
	}
	if err != nil {
		return err
	}
	if m.Role == "viewer" {
		return ErrViewerReadOnly
	}
	return nil
}

func withDefault(value, fallback string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return fallback
	}
	return v
}

func parseDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*raw)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, ValidationError{"due_date must be in YYYY-MM-DD format"}
	}
	return &t, nil
}