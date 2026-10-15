package entries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/nuvora-core/nuvora/internal/models"
)

// BrowseEntries lists the files and folders of a directory for the
// project file explorer. Without a path query parameter the user's
// home directory is used.
func BrowseEntries(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")

	if path == "" {
		home, err := os.UserHomeDir()

		if err != nil {
			fmt.Println("entries: home directory error:", err)

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

	dirEntries, err := os.ReadDir(abs)

	if err != nil {
		fmt.Println("entries: read directory error:", err)

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "cannot read directory",
		})

		return
	}

	entries := make([]models.Entry, 0, len(dirEntries))

	for _, entry := range dirEntries {
		isDir := entry.IsDir()

		if !isDir && entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Stat(filepath.Join(abs, entry.Name()))

			if err == nil && target.IsDir() {
				isDir = true
			}
		}

		entries = append(entries, models.Entry{
			Name:  entry.Name(),
			Path:  filepath.Join(abs, entry.Name()),
			IsDir: isDir,
		})
	}

	// Folders first, then files, both alphabetically.
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}

		return entries[i].Name < entries[j].Name
	})

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"path":    abs,
		"entries": entries,
	})
}
