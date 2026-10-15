package folders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuvora-core/nuvora/internal/models"
)

// BrowseFolders lists the subfolders of a directory so the user can
// pick a project root path. Without a path query parameter the user's
// home directory is used.
func BrowseFolders(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")

	if path == "" {
		home, err := os.UserHomeDir()

		if err != nil {
			fmt.Println("folders: home directory error:", err)

			w.WriteHeader(http.StatusInternalServerError)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "cannot determine home directory",
			})

			return
		}

		path = home
	}

	abs, err := filepath.Abs(path)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid path",
		})

		return
	}

	info, err := os.Stat(abs)

	if err != nil || !info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "not a directory",
		})

		return
	}

	entries, err := os.ReadDir(abs)

	if err != nil {
		fmt.Println("folders: read directory error:", err)

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "cannot read directory",
		})

		return
	}

	// os.ReadDir already returns the entries sorted by name.
	folders := []models.Folder{}

	for _, entry := range entries {
		// Hidden folders are usually not project roots.
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		isDir := entry.IsDir()

		if !isDir && entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Stat(filepath.Join(abs, entry.Name()))

			if err == nil && target.IsDir() {
				isDir = true
			}
		}

		if !isDir {
			continue
		}

		folders = append(folders, models.Folder{
			Name: entry.Name(),
			Path: filepath.Join(abs, entry.Name()),
		})
	}

	parent := filepath.Dir(abs)

	if parent == abs {
		// At the filesystem root there is nothing above.
		parent = ""
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"path":    abs,
		"parent":  parent,
		"folders": folders,
	})
}
