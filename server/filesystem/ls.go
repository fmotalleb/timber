package filesystem

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"

	"github.com/fmotalleb/timber/server/auth"
	"github.com/fmotalleb/timber/server/response"
)

type Node struct {
	Name     string  `json:"name"`
	Path     string  `json:"path,omitempty"`
	Type     string  `json:"type"`
	Children []*Node `json:"children,omitempty"`
	Size     int64   `json:"size"`
}

func (n *Node) getSize() int64 {
	if n.Type == "file" {
		return n.Size
	}
	if n.Size != 0 {
		return n.Size
	}
	for _, c := range n.Children {
		n.Size += c.getSize()
	}
	return n.Size
}

func (n *Node) findOrCreateChild(name string, nodeType string, size int64) *Node {
	for _, child := range n.Children {
		if child.Name == name {
			// A node can't be both a file and a directory.
			// If a directory is requested but a file exists, we can't proceed.
			// This indicates a conflict in the file structure or glob patterns.
			// For simplicity, we just return the existing node. The frontend can handle it.
			return child
		}
	}
	newNode := &Node{
		Name: name,
		Type: nodeType,
		Size: size,
	}
	n.Children = append(n.Children, newNode)
	sort.Slice(n.Children, func(i, j int) bool {
		if n.Children[i].Type != n.Children[j].Type {
			return n.Children[i].Type == "dir" // dirs first
		}
		return n.Children[i].Name < n.Children[j].Name
	})
	return newNode
}

func insertPath(root *Node, path string, isDir bool) {
	parts := strings.Split(path, string(os.PathSeparator))
	currentNode := root
	size := int64(0)
	if stat, err := os.Stat(path); err == nil {
		size = stat.Size()
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		nodeType := "dir"
		// The last part of the path determines the type
		fSize := int64(0)
		if i == len(parts)-1 && !isDir {
			nodeType = "file"
			fSize = size
		}
		currentNode = currentNode.findOrCreateChild(part, nodeType, fSize)
		// Set the full path only for the final node in the path
		if i == len(parts)-1 {
			currentNode.Path = path
		}
	}
}

// Ls returns a list of files that the user has access to.
func Ls(w http.ResponseWriter, r *http.Request) {
	access, ok := auth.AccessFromContext(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	logger := log.Of(r.Context())

	root := &Node{Name: "root", Type: "dir"}
	processed := make(map[string]bool)

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
			// Clean the path to have consistent separators
			cleanedPath := filepath.ToSlash(match)
			if processed[cleanedPath] {
				continue
			}

			stat, err := os.Stat(cleanedPath)
			if err != nil {
				logger.Warn("failed to stat file", zap.String("path", cleanedPath), zap.Error(err))
				continue
			}
			insertPath(root, cleanedPath, stat.IsDir())
			processed[cleanedPath] = true
		}
	}
	root.getSize()
	if err := response.JSON(w, root.Children, http.StatusOK); err != nil {
		logger.Error("failed to write response", zap.Error(err))
	}
}
