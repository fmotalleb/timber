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

	lines := getLinesParam(r, defaultLineCount)

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	reader := bufio.NewReader(f)

	for i := 0; i < lines; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				w.Write([]byte(line))
			} else if err != io.EOF {
				logger.Error("failed to read file", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			break
		}
		w.Write([]byte(line))
	}
}
