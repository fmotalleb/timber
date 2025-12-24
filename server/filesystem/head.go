package filesystem

import (
	"bufio"
	"io"
	"net/http"
	"os"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/helper"
)

// Head returns the first n lines of a file.
func Head(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.Of(ctx)

	filePath, ok := helper.GetPath(r)
	if !ok {
		http.Error(w, "missing `path` query parameter", http.StatusBadRequest)
		return
	}
	if containsDotDot(filePath) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	lines := getLinesParam(r, defaultLineCount)
	if lines <= 0 {
		http.Error(w, "`lines` must be greater than zero", http.StatusBadRequest)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			http.Error(w, "file not found", http.StatusNotFound)
		case os.IsPermission(err):
			http.Error(w, "permission denied", http.StatusForbidden)
		default:
			logger.Error("failed to open file", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			logger.Warn("failed to close file", zap.Error(cerr))
		}
	}()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	reader := bufio.NewReader(f)

	for i := 0; i < lines; i++ {
		select {
		case <-ctx.Done():
			logger.Warn("request canceled", zap.Error(ctx.Err()))
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					if _, werr := w.Write([]byte(line)); werr != nil {
						logger.Warn("failed to write response", zap.Error(werr))
					}
				}
				return
			}

			logger.Error("failed to read file", zap.Error(err))
			return
		}

		if _, werr := w.Write([]byte(line)); werr != nil {
			logger.Warn("failed to write response", zap.Error(werr))
			return
		}
	}
}
