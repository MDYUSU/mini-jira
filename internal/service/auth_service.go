package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/MDYUSU/mini-jira/internal/auth"
	"github.com/MDYUSU/mini-jira/internal/model"
	"github.com/MDYUSU/mini-jira/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type RegisterInput struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      model.User `json:"user"`
}

type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (AuthResponse, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return AuthResponse{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return AuthResponse{}, ValidationError{"name is required"}
	}
	if len(in.Password) < 8 || len(in.Password) > 72 {
		return AuthResponse{}, ValidationError{"password must be 8 to 72 characters"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.users.CreateWithPassword(ctx, email, name, string(hash))
	if err != nil {
		return AuthResponse{}, err
	}
	return s.respond(user)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return AuthResponse{}, ValidationError{"email and password are required"}
	}

	creds, err := s.users.GetCredentials(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if creds.PasswordHash == nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(*creds.PasswordHash), []byte(in.Password)) != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	user, err := s.users.GetBasic(ctx, creds.UserID)
	if err != nil {
		return AuthResponse{}, err
	}
	return s.respond(user)
}

func (s *AuthService) respond(user model.User) (AuthResponse, error) {
	token, expiresAt, err := s.tokens.Issue(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, ExpiresAt: expiresAt, User: user}, nil
}
