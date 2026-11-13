package service

import (
	"context"
	"errors"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

var ErrNotAuthor = errors.New("only the author can edit this comment")

var ErrCannotDeleteComment = errors.New("only the author or a project owner can delete this comment")

type CreateCommentInput struct {
	AuthorID int64  `json:"author_id"`
	Body     string `json:"body"`
}

type UpdateCommentInput struct {
	AuthorID int64  `json:"author_id"`
	Body     string `json:"body"`
}

type CommentService struct {
	comments *repository.CommentRepository
	tasks    *repository.TaskRepository
	projects *repository.ProjectRepository
}

func NewCommentService(comments *repository.CommentRepository, tasks *repository.TaskRepository, projects *repository.ProjectRepository) *CommentService {
	return &CommentService{comments: comments, tasks: tasks, projects: projects}
}

func (s *CommentService) ListByTask(ctx context.Context, taskID int64) ([]model.Comment, error) {
	exists, err := s.tasks.Exists(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, repository.ErrNotFound
	}
	return s.comments.ListByTask(ctx, taskID)
}

func (s *CommentService) Create(ctx context.Context, taskID int64, in CreateCommentInput) (model.Comment, error) {
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return model.Comment{}, ValidationError{"body is required"}
	}
	if in.AuthorID < 1 {
		return model.Comment{}, ValidationError{"author_id is required"}
	}

	projectID, err := s.tasks.GetProjectID(ctx, taskID)
	if err != nil {
		return model.Comment{}, err
	}
	if _, err := checkMember(ctx, s.projects, projectID, in.AuthorID, "author"); err != nil {
		return model.Comment{}, err
	}

	id, err := s.comments.Create(ctx, taskID, in.AuthorID, body)
	if err != nil {
		return model.Comment{}, err
	}
	return s.comments.GetByID(ctx, id)
}

func (s *CommentService) Update(ctx context.Context, id int64, in UpdateCommentInput) (model.Comment, error) {
	body := strings.TrimSpace(in.Body)
	if body == "" {
		return model.Comment{}, ValidationError{"body is required"}
	}
	if in.AuthorID < 1 {
		return model.Comment{}, ValidationError{"author_id is required"}
	}

	info, err := s.comments.GetInfo(ctx, id)
	if err != nil {
		return model.Comment{}, err
	}
	if info.AuthorID == nil || *info.AuthorID != in.AuthorID {
		return model.Comment{}, ErrNotAuthor
	}

	if err := s.comments.UpdateBody(ctx, id, body); err != nil {
		return model.Comment{}, err
	}
	return s.comments.GetByID(ctx, id)
}

func (s *CommentService) Delete(ctx context.Context, id, actorID int64) error {
	info, err := s.comments.GetInfo(ctx, id)
	if err != nil {
		return err
	}

	if info.AuthorID != nil && *info.AuthorID == actorID {
		return s.comments.Delete(ctx, id)
	}

	m, err := s.projects.GetMember(ctx, info.ProjectID, actorID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrCannotDeleteComment
	}
	if err != nil {
		return err
	}
	if m.Role != "owner" {
		return ErrCannotDeleteComment
	}
	return s.comments.Delete(ctx, id)
}