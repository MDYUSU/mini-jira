package service

import (
	"context"

	"github.com/MDYUSU/mini-jira/internal/model"
)

func (s *AuthService) Me(ctx context.Context, userID int64) (model.UserWithProfile, error) {
	return s.users.GetByID(ctx, userID)
}
