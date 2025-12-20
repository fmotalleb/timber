package filesystem

import (
	"net/http"
	"path/filepath"
	"sort"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/auth"
	"github.com/fmotalleb/timber/server/response"
)

// Ls returns a list of files that the user has access to.
func Ls(w http.ResponseWriter, r *http.Request) {
	access, ok := auth.AccessFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	logger := log.Of(r.Context())
	fileSet := make(map[string]struct{})
	for _, pat := range access {
		matches, patErr := filepath.Glob(pat)
		if patErr != nil {
			logger.Error(
				"failed to parse glob pattern",
				zap.String("pattern", pat),
				zap.Error(patErr),
			)
			continue
		}
		for _, match := range matches {
			fileSet[match] = struct{}{}
		}
	}

	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}
	sort.Strings(files)

	if err := response.JSON(w, files, http.StatusOK); err != nil {
		logger.Error("failed to write response", zap.Error(err))
	}
}
