package filesystem

import (
	"io"
	"net/http"
	"os"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/helper"
)

func Tail(w http.ResponseWriter, r *http.Request) {
	filePath, ok := helper.GetPath(r)
	if !ok {
		http.Error(w, "missing `path` query parameter", http.StatusBadRequest)
		return
	}

	lines := getLinesParam(r, 10)
	follow := getFollowParam(r)

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	last, err := tailLines(f, lines)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, l := range last {
		if l != "" {
			if _, err := w.Write([]byte(l)); err != nil {
				log.Of(r.Context()).Error("failed to write response", zap.Error(err))
			}
		}
	}

	if follow {
		_, _ = f.Seek(0, io.SeekEnd)
		followFile(w, r, f)
	}
}
