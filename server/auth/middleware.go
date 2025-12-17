package auth

import (
	"context"
	"net/http"

	"github.com/fmotalleb/go-tools/log"

	"github.com/fmotalleb/timber/config"
	"github.com/fmotalleb/timber/server/response"
)

func WithBasicAuth(cfg config.Config) func(http.Handler) http.Handler {
	// build user index once
	users := make(map[string]config.User, len(cfg.Users))
	for _, u := range cfg.Users {
		users[u.Name] = u
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.Of(r.Context())
			username, password, ok := r.BasicAuth()
			if !ok {
				logger.Warn("no auth found")
				response.Unauthorized(w)
				return
			}

			u, ok := users[username]
			if !ok || u.Password != password {
				logger.Warn("authentication failed")
				response.Unauthorized(w)
				return
			}

			// resolve access lists
			var access []string
			for _, name := range u.AccessList {
				a, ok := cfg.Access[name]
				if !ok {
					continue
				}
				access = append(access, a.Paths...)
			}

			authUser := &AuthUser{
				Name:   u.Name,
				Access: access,
			}

			ctx := context.WithValue(r.Context(), ctxUserKey, authUser)
			ctx = context.WithValue(ctx, ctxAccessKey, access)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
