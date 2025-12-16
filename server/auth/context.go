package auth

import (
	"context"
)

type ctxKey string

const (
	ctxUserKey   ctxKey = "auth.user"
	ctxAccessKey ctxKey = "auth.access"
)

type AuthUser struct {
	Name   string   `json:"name"`
	Access []string `json:"access"`
}

func UserFromContext(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(ctxUserKey).(*AuthUser)
	return u, ok
}

func AccessFromContext(ctx context.Context) ([]string, bool) {
	a, ok := ctx.Value(ctxAccessKey).([]string)
	return a, ok
}
