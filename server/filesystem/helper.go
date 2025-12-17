package filesystem

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

const (
	readChunkSize          = 4096
	followFilePollInterval = 200 * time.Millisecond
	defaultLineCount       = 10
)

func getLinesParam(r *http.Request, def int) int {
	v := r.URL.Query().Get("lines")
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	return def
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.Split(v, "/") {
		if ent == ".." {
			return true
		}
	}
	for _, ent := range strings.Split(v, `\`) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func getFollowParam(r *http.Request) bool {
	v := strings.ToLower(r.URL.Query().Get("follow"))
	return v == "1" || v == "true" || v == "yes"
}

func tailLines(f *os.File, n int) ([]string, error) {
	const blockSize = readChunkSize

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	var (
		buf    []byte
		lines  []string
		offset = size
	)

	for offset > 0 && len(lines) <= n {
		readSize := int64(blockSize)
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		tmp := make([]byte, readSize)
		if _, err := f.ReadAt(tmp, offset); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		buf = append(tmp, buf...)
		lines = strings.Split(string(buf), "\n")
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}

func followFile(w http.ResponseWriter, r *http.Request, f *os.File) {
	h := w.Header()
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()

	reader := bufio.NewReader(f)
	ctx := r.Context()
	buf := make([]byte, readChunkSize) // read chunks of 4KB

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				log.Of(r.Context()).Error("failed to write response", zap.Error(writeErr))
			}
			flusher.Flush()
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(followFilePollInterval)
				continue
			}
			return
		}
	}
}
