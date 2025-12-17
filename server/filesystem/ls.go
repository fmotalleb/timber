package filesystem

import (
	"net/http"
	"path/filepath"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/auth"
	"github.com/fmotalleb/timber/server/response"
)

func Ls(w http.ResponseWriter, r *http.Request) {
	access, ok := auth.AccessFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	logger := log.Of(r.Context())
	files := make([]string, 0)
	for _, pat := range access {
		matches, patErr := filepath.Glob(pat)
		if patErr != nil {
			logger.Error(
				"failed to parse glob pattern",
				zap.String("pattern", pat),
				zap.Error(patErr),
			)
		}
		files = append(files, matches...)
	}
	if err := response.Json(w, files, http.StatusOK); err != nil {
		logger.Error("failed to write response", zap.Error(err))
	}
}
