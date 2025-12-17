package filesystem

import (
	"bufio"
	"net/http"
	"os"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/helper"
)

// Head returns the first n lines of a file.
func Head(w http.ResponseWriter, r *http.Request) {
	logger := log.Of(r.Context())
	filePath, ok := helper.GetPath(r)
	if !ok {
		http.Error(w, "missing `path` query parameter", http.StatusBadRequest)
		return
	}
	if containsDotDot(filePath) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	lines := getLinesParam(r, DefaultLineCount)

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	sc := bufio.NewScanner(f)
	for i := 0; i < lines && sc.Scan(); i++ {
		if _, err := w.Write(append(sc.Bytes(), '\n')); err != nil {
			logger.Error("failed to write response", zap.Error(err))
		}
	}

	if err := sc.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
