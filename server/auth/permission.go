package auth

import (
	"errors"
	"net/http"
	"path"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/helper"
	"github.com/fmotalleb/timber/server/response"
)

var ErrorPermissionDeny = errors.New("permission denied")

func PermissionCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if err := permissionHandler(w, r); err != nil {
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}

func permissionHandler(w http.ResponseWriter, r *http.Request) error {
	logger := log.Of(r.Context())
	reqPath, ok := helper.GetPath(r)
	if !ok {
		return nil
	}
	access, ok := AccessFromContext(r.Context())
	if !ok {
		logger.Warn("user not found in the request context")
		response.PermissionDenied(w)
		return ErrorPermissionDeny
	}
	for _, acc := range access {
		if matched, err := path.Match(acc, reqPath); err == nil && matched {
			return nil
		} else if err != nil {
			logger.Warn("path match evaluation failed", zap.Error(err))
		}
	}

	response.PermissionDenied(w)
	return ErrorPermissionDeny
}
