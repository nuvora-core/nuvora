package tools

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// Resolve turns a project-relative path into an absolute path and
// verifies it stays inside the project root. Both relative and
// absolute paths are accepted, as long as they point inside the root.
// Symlinks are followed to make sure they cannot escape the root.
func Resolve(projectRoot, path string) (string, error) {
	root, err := filepath.Abs(projectRoot)

	if err != nil {
		return "", errors.New("invalid project root")
	}

	if path == "" {
		path = "."
	}

	target := path

	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}

	target = filepath.Clean(target)

	if !isInside(root, target) {
		return "", errors.New("path is outside the project root")
	}

	// Follow symlinks on the deepest existing ancestor so a link
	// pointing outside the root cannot be used to escape it.
	ancestor, ok := resolvedExistingAncestor(target)

	if ok && !isInside(root, ancestor) {
		return "", errors.New("path is outside the project root")
	}

	return target, nil
}

// isInside reports whether path is root itself or below it.
func isInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)

	if err != nil {
		return false
	}

	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolvedExistingAncestor resolves symlinks on the deepest existing
// ancestor of path. It reports false when nothing on the path exists.
func resolvedExistingAncestor(path string) (string, bool) {
	current := path

	for {
		resolved, err := filepath.EvalSymlinks(current)

		if err == nil {
			return resolved, true
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", false
		}

		parent := filepath.Dir(current)

		if parent == current {
			return "", false
		}

		current = parent
	}
}
