package service

import (
	"context"
	"errors"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

func checkMember(ctx context.Context, projects *repository.ProjectRepository, projectID, userID int64, who string) (model.ProjectMember, error) {
	m, err := projects.GetMember(ctx, projectID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return m, ValidationError{who + " must be a member of this project"}
	}
	return m, err
}

func checkEditor(ctx context.Context, projects *repository.ProjectRepository, projectID, userID int64, who string) error {
	m, err := checkMember(ctx, projects, projectID, userID, who)
	if err != nil {
		return err
	}
	if m.Role == "viewer" {
		return ErrViewerReadOnly
	}
	return nil
}