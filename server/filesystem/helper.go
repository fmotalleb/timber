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
	const blockSize = 4096 // Size of each chunk to read (adjust as needed)

	// Get file information (size, etc.)
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	var (
		lines    []string // Stores the last n lines
		offset   = size   // Start from the end of the file
		buf      = make([]byte, blockSize)
		lastLine []byte // To store an incomplete line across chunks
	)

	// Loop to read chunks from the end of the file
	for offset > 0 && len(lines) <= n+1 {
		// Determine how much to read in this iteration
		readSize := blockSize
		if offset < int64(readSize) {
			readSize = int(offset)
		}

		// Move the offset backwards
		offset -= int64(readSize)

		// Read the chunk from the file
		_, err := f.ReadAt(buf[:readSize], offset)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		// Append the incomplete last line (if any) from the previous chunk
		if len(lastLine) > 0 {
			buf = append(lastLine, buf[:readSize]...) // Prepend the last line to the current chunk
			readSize += len(lastLine)                 // Adjust the read size
			lastLine = nil                            // Reset for next chunk
		}

		// Split the chunk into lines
		chunkLines := strings.Split(string(buf[:readSize]), "\n")

		// If the last chunk ends with an incomplete line, store it
		if !strings.HasSuffix(string(buf[:readSize]), "\n") {
			lastLine = []byte(chunkLines[len(chunkLines)-1])
			chunkLines = chunkLines[:len(chunkLines)-1] // Remove the incomplete line
		}

		// Prepend the chunk lines to the lines slice (we do this in reverse order)
		lines = append(chunkLines, lines...)

		// If we now have more than n lines, trim the excess
		if len(lines) > n+1 {
			lines = lines[:n+1]
		}
	}

	// Handle the case where the file ends without a newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // Remove the empty last line (if present)
	}
	if len(lines) > 1 {
		lines = lines[1:]
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
