package service

import (
	"context"
	"math"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

func (s *ProjectService) Stats(ctx context.Context, projectID int64) (model.ProjectStats, error) {
	if err := s.requireProject(ctx, projectID); err != nil {
		return model.ProjectStats{}, err
	}
	stats, err := s.projects.Stats(ctx, projectID)
	if err != nil {
		return stats, err
	}
	if stats.Total > 0 {
		stats.DonePercent = math.Round(float64(stats.ByStatus.Done)*1000/float64(stats.Total)) / 10
	}
	return stats, nil
}

func (s *ProjectService) Workload(ctx context.Context, projectID int64) ([]model.MemberWorkload, error) {
	if err := s.requireProject(ctx, projectID); err != nil {
		return nil, err
	}
	return s.projects.Workload(ctx, projectID)
}

func (s *ProjectService) requireProject(ctx context.Context, projectID int64) error {
	exists, err := s.projects.Exists(ctx, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrNotFound
	}
	return nil
}
