package server

import (
	"context"

	"athenaeum/internal/models"
)

type ctxKey int

const userContextKey ctxKey = 1

// WithUser attaches a user to the request context.
func WithUser(ctx context.Context, u models.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext returns the authenticated user and whether one was set.
func UserFromContext(ctx context.Context) (models.User, bool) {
	u, ok := ctx.Value(userContextKey).(models.User)
	return u, ok
}

// UserIDFromContext returns the authenticated user id or AnonymousUserID.
func UserIDFromContext(ctx context.Context) int64 {
	if u, ok := UserFromContext(ctx); ok {
		return u.ID
	}
	return models.AnonymousUserID
}
