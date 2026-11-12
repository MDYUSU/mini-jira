package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UpdateUserInput struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
}

type ProfileInput struct {
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) Get(ctx context.Context, id int64) (model.UserWithProfile, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (model.User, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return model.User{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.User{}, ValidationError{"name is required"}
	}
	return s.users.Create(ctx, email, name)
}

func (s *UserService) Update(ctx context.Context, id int64, in UpdateUserInput) (model.User, error) {
	if in.Email == nil && in.Name == nil {
		return model.User{}, ValidationError{"nothing to update"}
	}
	if in.Email != nil {
		email, err := normalizeEmail(*in.Email)
		if err != nil {
			return model.User{}, err
		}
		in.Email = &email
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return model.User{}, ValidationError{"name cannot be empty"}
		}
		in.Name = &name
	}
	return s.users.Update(ctx, id, in.Email, in.Name)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.users.Delete(ctx, id)
}

func (s *UserService) UpsertProfile(ctx context.Context, id int64, in ProfileInput) (model.UserProfile, error) {
	p, err := s.users.UpsertProfile(ctx, id, in.Bio, in.AvatarURL)
	if errors.Is(err, repository.ErrForeignKey) {
		return p, repository.ErrNotFound
	}
	return p, err
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", ValidationError{"email is required"}
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ValidationError{"email is invalid"}
	}
	return email, nil
}