package auth

import "context"

type contextKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, contextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(contextKey{}).(int64)
	return id, ok && id > 0
}
