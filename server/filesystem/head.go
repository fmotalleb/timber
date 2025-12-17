package filesystem

import (
	"bufio"
	"net/http"
	"os"

	"github.com/fmotalleb/timber/server/helper"
)

func Head(w http.ResponseWriter, r *http.Request) {
	filePath, ok := helper.GetPath(r)
	if !ok {
		http.Error(w, "missing `path` query parameter", http.StatusBadRequest)
		return
	}

	lines := getLinesParam(r, 10)

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	sc := bufio.NewScanner(f)
	for i := 0; i < lines && sc.Scan(); i++ {
		w.Write(append(sc.Bytes(), '\n'))
	}

	if err := sc.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
