package auth

import (
	"context"
)

type ctxKey string

const (
	ctxUserKey   ctxKey = "auth.user"
	ctxAccessKey ctxKey = "auth.access"
)

// AuthUser represents the authenticated user.
type AuthUser struct {
	Name   string   `json:"name"`
	Access []string `json:"access"`
}

// UserFromContext returns the authenticated user from the context.
func UserFromContext(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(ctxUserKey).(*AuthUser)
	return u, ok
}

// AccessFromContext returns the access rights of the authenticated user from the context.
func AccessFromContext(ctx context.Context) ([]string, bool) {
	a, ok := ctx.Value(ctxAccessKey).([]string)
	return a, ok
}
